"""LangGraph + Azure OpenAI streaming server with Prometheus metrics."""
from __future__ import annotations

import asyncio
import os
import time
from contextlib import asynccontextmanager
from typing import AsyncIterator

from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse, Response
from langchain_core.messages import HumanMessage, SystemMessage
from langchain_openai import AzureChatOpenAI
from langgraph.graph import START, END, StateGraph
from prometheus_client import (
    CONTENT_TYPE_LATEST,
    Counter,
    Gauge,
    Histogram,
    generate_latest,
)
from pydantic import BaseModel
from typing_extensions import TypedDict

MAX_CONCURRENCY = int(os.environ.get("MAX_CONCURRENCY", "128"))

ttft_seconds = Histogram(
    "ttft_seconds",
    "Time to first token at the API edge.",
    ["route", "mode"],
    buckets=(0.1, 0.25, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 5.0, 10.0),
)
inter_token_latency_seconds = Histogram(
    "inter_token_latency_seconds",
    "Inter-token gap during SSE streaming.",
    ["route"],
    buckets=(0.01, 0.02, 0.04, 0.06, 0.08, 0.12, 0.2, 0.4, 0.8, 1.6),
)
total_response_seconds = Histogram(
    "total_response_seconds",
    "Request to last token wall-clock time.",
    ["route", "status"],
    buckets=(0.5, 1.0, 2.0, 3.0, 5.0, 8.0, 10.0, 15.0, 30.0, 60.0),
)
llm_requests_total = Counter(
    "llm_requests_total", "Total LLM requests.", ["route", "status"]
)
llm_errors_total = Counter(
    "llm_errors_total", "LLM error counter.", ["route", "kind"]
)
request_queue_depth = Gauge(
    "request_queue_depth", "Requests waiting on the LLM concurrency semaphore."
)
concurrent_llm_calls = Gauge(
    "concurrent_llm_calls", "In-flight LLM calls holding a semaphore slot."
)
concurrent_sessions = Gauge(
    "concurrent_sessions", "Active SSE sessions at the HTTP layer."
)


class GraphState(TypedDict):
    messages: list


def _build_graph():
    llm = AzureChatOpenAI(
        azure_endpoint=os.environ["AZURE_AI_FOUNDRY_ENDPOINT"],
        api_key=os.environ["AZURE_AI_FOUNDRY_API_KEY"],
        azure_deployment=os.environ["AZURE_AI_FOUNDRY_DEPLOYMENT"],
        api_version=os.environ.get("AZURE_AI_FOUNDRY_API_VERSION", "2024-10-21"),
        streaming=True,
        temperature=0.2,
    )

    async def call_model(state: GraphState) -> GraphState:
        msgs = [SystemMessage(content="Answer briefly.")] + [
            HumanMessage(content=m["content"]) for m in state["messages"]
        ]
        result = await llm.ainvoke(msgs)
        return {"messages": state["messages"] + [{"role": "assistant", "content": result.content}]}

    g = StateGraph(GraphState)
    g.add_node("model", call_model)
    g.add_edge(START, "model")
    g.add_edge("model", END)
    return g.compile(), llm


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    app.state.semaphore = asyncio.Semaphore(MAX_CONCURRENCY)
    app.state.graph, app.state.llm = _build_graph()
    app.state.queue_depth = 0
    yield


app = FastAPI(lifespan=lifespan)


class ChatMessage(BaseModel):
    role: str
    content: str


class ChatRequest(BaseModel):
    messages: list[ChatMessage]


@app.get("/healthz")
async def healthz() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/metrics")
async def metrics() -> Response:
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)


@app.post("/chat")
async def chat(req: ChatRequest, request: Request) -> StreamingResponse:
    route = "/chat"
    mode = "stream"
    sem: asyncio.Semaphore = request.app.state.semaphore
    llm = request.app.state.llm
    msgs = [SystemMessage(content="Answer briefly.")] + [
        HumanMessage(content=m.content) for m in req.messages
    ]

    async def event_stream() -> AsyncIterator[bytes]:
        concurrent_sessions.inc()
        request.app.state.queue_depth += 1
        request_queue_depth.set(request.app.state.queue_depth)
        start = time.perf_counter()
        first_token_at: float | None = None
        last_token_at: float | None = None
        status_label = "ok"
        try:
            async with sem:
                request.app.state.queue_depth -= 1
                request_queue_depth.set(request.app.state.queue_depth)
                concurrent_llm_calls.inc()
                try:
                    async for chunk in llm.astream(msgs):
                        text = getattr(chunk, "content", "") or ""
                        if not text:
                            continue
                        now = time.perf_counter()
                        if first_token_at is None:
                            first_token_at = now
                            ttft_seconds.labels(route, mode).observe(now - start)
                        else:
                            inter_token_latency_seconds.labels(route).observe(
                                now - (last_token_at or now)
                            )
                        last_token_at = now
                        yield f"data: {text}\n\n".encode()
                    yield b"data: [DONE]\n\n"
                finally:
                    concurrent_llm_calls.dec()
        except Exception as exc:
            status_label = "error"
            llm_errors_total.labels(route, type(exc).__name__).inc()
            yield f"data: [ERROR] {type(exc).__name__}\n\n".encode()
        finally:
            elapsed = time.perf_counter() - start
            total_response_seconds.labels(route, status_label).observe(elapsed)
            llm_requests_total.labels(route, status_label).inc()
            concurrent_sessions.dec()
            if first_token_at is None and request.app.state.queue_depth > 0:
                request.app.state.queue_depth -= 1
                request_queue_depth.set(request.app.state.queue_depth)

    return StreamingResponse(event_stream(), media_type="text/event-stream")
