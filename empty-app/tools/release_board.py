"""Agent Sail rich.Live release board.

Layout matches docs/02-agent-sail/proposal.md section 7:

  Customers | Criteria / Evidence | Verdict
  Live Load Probe (Phase/Users/RPS, p50/p95/error, sparklines)
"""
from __future__ import annotations

import collections
import os
import re
import subprocess
import sys
import threading
import time
import urllib.request

import yaml
from rich.console import Console
from rich.layout import Layout
from rich.live import Live
from rich.panel import Panel
from rich.table import Table
from rich.text import Text

CONFIG_PATH = sys.argv[1] if len(sys.argv) > 1 else "agentsail.loadtest.yaml"
CUSTOMER = sys.argv[2] if len(sys.argv) > 2 else "demo"

cfg = yaml.safe_load(open(CONFIG_PATH))
target_url = cfg["target"]["base_url"] + cfg["target"]["chat_path"]
metrics_url = cfg["target"]["metrics_url"]
locust_cfg = cfg["locust"]
slos = cfg["slos"]
runtime = cfg.get("runtime", {})

state = {
    "phase": "READY",
    "started_at": None,
    "requests": 0,
    "errors": 0,
    "rps_samples": collections.deque(maxlen=60),
    "ttft_p95_samples": collections.deque(maxlen=60),
    "total_p50_samples": collections.deque(maxlen=60),
    "total_p95_samples": collections.deque(maxlen=60),
    "users_current": 0,
    "verdict": "RUNNING",
    "fail_reasons": [],
}

console = Console()


def parse_metrics(text: str) -> dict[str, list[tuple[str, float]]]:
    out: dict[str, list[tuple[str, float]]] = {}
    pat = re.compile(r"^(\w+)(?:\{([^}]*)\})?\s+([\d.eE+\-]+)$")
    for line in text.splitlines():
        if line.startswith("#") or not line.strip():
            continue
        m = pat.match(line)
        if not m:
            continue
        name, labels, val = m.group(1), m.group(2) or "", float(m.group(3))
        out.setdefault(name, []).append((labels, val))
    return out


def histogram_quantile(metrics: dict, name: str, q: float) -> float | None:
    buckets: list[tuple[float, float]] = []
    total = 0.0
    for labels, val in metrics.get(f"{name}_bucket", []):
        m = re.search(r'le="([^"]+)"', labels)
        if not m:
            continue
        le = m.group(1)
        if le == "+Inf":
            total = max(total, val)
            continue
        try:
            buckets.append((float(le), val))
        except ValueError:
            continue
    if not buckets:
        return None
    if total == 0:
        total = max(c for _, c in buckets)
    if total == 0:
        return None
    buckets.sort()
    target = q * total
    prev_le, prev_count = 0.0, 0.0
    for le, c in buckets:
        if c >= target:
            if c == prev_count:
                return le
            frac = (target - prev_count) / (c - prev_count)
            return prev_le + frac * (le - prev_le)
        prev_le, prev_count = le, c
    return buckets[-1][0]


def gauge_value(metrics: dict, name: str) -> float:
    return sum(v for _, v in metrics.get(name, []))


def sparkline(samples: list[float]) -> str:
    chars = "▁▂▃▄▅▆▇█"
    if not samples:
        return ""
    lo, hi = min(samples), max(samples)
    if hi == lo:
        return chars[0] * len(samples)
    return "".join(
        chars[int((s - lo) / (hi - lo) * (len(chars) - 1))] for s in samples
    )


def poll_metrics() -> None:
    last_req, last_t = 0.0, time.time()
    while state["phase"] != "DONE":
        try:
            with urllib.request.urlopen(metrics_url, timeout=2) as r:
                m = parse_metrics(r.read().decode())
            req = gauge_value(m, "llm_requests_total")
            err = gauge_value(m, "llm_errors_total")
            now = time.time()
            dt = now - last_t
            if dt > 0:
                state["rps_samples"].append((req - last_req) / dt)
            state["requests"] = int(req)
            state["errors"] = int(err)
            state["users_current"] = int(gauge_value(m, "concurrent_sessions"))
            for hist, key, q in (
                ("ttft_seconds", "ttft_p95_samples", 0.95),
                ("total_response_seconds", "total_p50_samples", 0.5),
                ("total_response_seconds", "total_p95_samples", 0.95),
            ):
                v = histogram_quantile(m, hist, q)
                if v is not None:
                    state[key].append(v)
            last_req, last_t = req, now
        except Exception:
            pass
        time.sleep(1)


def run_locust() -> None:
    state["started_at"] = time.time()
    state["phase"] = "RAMPING"
    venv = runtime.get("venv", ".agentsail/loadtests/.venv")
    results = locust_cfg["results_dir"]
    os.makedirs(results, exist_ok=True)
    cmd = [
        f"{venv}/bin/locust",
        "-f", locust_cfg["locustfile"],
        "--host", cfg["target"]["base_url"],
        "--headless",
        "--users", str(locust_cfg["users"]),
        "--spawn-rate", str(locust_cfg["spawn_rate"]),
        "--run-time", str(locust_cfg["run_time"]),
        "--csv", f"{results}/stats",
        "--html", f"{results}/report.html",
    ]
    p = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    for line in p.stdout or []:
        low = line.lower()
        if "all users spawned" in low or "all users have been spawned" in low:
            state["phase"] = "STEADY"
        if "shutting down" in low:
            state["phase"] = "STOPPING"
    p.wait()
    state["phase"] = "DONE"
    fails: list[str] = []
    if state["ttft_p95_samples"]:
        v = max(state["ttft_p95_samples"])
        thr = slos["ttft_p95_seconds"]["threshold"]
        if v > thr:
            fails.append(f"ttft p95 {v:.2f}s > {thr}s")
    if state["total_p95_samples"]:
        v = max(state["total_p95_samples"])
        thr = slos["total_response_p95_seconds"]["threshold"]
        if v > thr:
            fails.append(f"total p95 {v:.2f}s > {thr}s")
    if state["requests"]:
        er = state["errors"] / state["requests"]
        thr = slos["error_rate"]["threshold"]
        if er > thr:
            fails.append(f"error rate {er:.2%} > {thr:.0%}")
    state["fail_reasons"] = fails
    state["verdict"] = "BLOCK" if fails else "SHIP"


def panel_customers() -> Panel:
    body = Text()
    body.append(CUSTOMER + "\n", style="bold cyan")
    body.append("\n")
    body.append("Target\n", style="bold")
    body.append(target_url + "\n")
    body.append(f"\nProfile\n", style="bold")
    body.append(f"{locust_cfg['users']} users / {locust_cfg['run_time']}\n")
    body.append(f"spawn {locust_cfg['spawn_rate']}/s\n")
    body.append(f"\nMode\n", style="bold")
    body.append(os.environ.get("LLM_MODE", "real"))
    return Panel(body, title="Customers", border_style="cyan")


def panel_criteria() -> Panel:
    body = Text()
    items = [
        ("ttft_p95_seconds", f"<= {slos['ttft_p95_seconds']['threshold']}s"),
        ("inter_token_latency_p95_seconds", f"<= {slos['inter_token_latency_p95_seconds']['threshold']}s"),
        ("total_response_p95_seconds", f"<= {slos['total_response_p95_seconds']['threshold']}s"),
        ("error_rate", f"<= {slos['error_rate']['threshold']:.0%}"),
        ("memory_working_set", f"<= {slos['memory_working_set']['threshold']} of limit"),
    ]
    for k, v in items:
        body.append(f"{k}: ", style="bold")
        body.append(f"{v}\n")
    body.append(f"\nsource: {os.path.basename(CONFIG_PATH)}\n", style="dim")
    return Panel(body, title="Criteria / Evidence", border_style="yellow")


def panel_verdict() -> Panel:
    color = {
        "RUNNING": "yellow",
        "RAMPING": "yellow",
        "STEADY": "blue",
        "STOPPING": "magenta",
        "SHIP": "green",
        "BLOCK": "red",
        "HOLD": "yellow",
    }.get(state["verdict"], "white")
    txt = Text()
    txt.append(state["verdict"] + "\n", style=f"bold {color}")
    if state["fail_reasons"]:
        txt.append("\nFailed:\n", style="bold red")
        for i, r in enumerate(state["fail_reasons"], 1):
            txt.append(f" {i}. {r}\n")
    else:
        txt.append("\nphase: ", style="bold")
        txt.append(state["phase"] + "\n")
    return Panel(txt, title="Verdict", border_style=color)


def panel_probe() -> Panel:
    rps = state["rps_samples"][-1] if state["rps_samples"] else 0.0
    p50 = state["total_p50_samples"][-1] if state["total_p50_samples"] else None
    p95 = state["total_p95_samples"][-1] if state["total_p95_samples"] else None
    er = state["errors"] / state["requests"] if state["requests"] else 0.0

    p50_str = f"{p50:.2f}s" if p50 else "-"
    p95_str = f"{p95:.2f}s" if p95 else "-"

    t = Table.grid(padding=(0, 4), expand=True)
    t.add_row(
        f"[bold]Phase:[/bold] {state['phase']}",
        f"[bold]Users:[/bold] {state['users_current']} / {locust_cfg['users']}",
        f"[bold]RPS:[/bold] {rps:.1f}",
    )
    t.add_row(
        f"[bold]p50:[/bold] {p50_str}",
        f"[bold]p95:[/bold] {p95_str}",
        f"[bold]error:[/bold] {er:.2%}",
    )
    t.add_row("")
    t.add_row(f"[cyan]Load   {sparkline(list(state['rps_samples']))}[/cyan]")
    t.add_row(f"[magenta]p95    {sparkline(list(state['total_p95_samples']))}[/magenta]")
    return Panel(t, title="Live Load Probe", border_style="green")


def render() -> Layout:
    layout = Layout()
    layout.split_column(
        Layout(name="top", size=14),
        Layout(name="probe"),
    )
    layout["top"].split_row(
        Layout(name="customers", ratio=1),
        Layout(name="criteria", ratio=2),
        Layout(name="verdict", ratio=1),
    )
    layout["customers"].update(panel_customers())
    layout["criteria"].update(panel_criteria())
    layout["verdict"].update(panel_verdict())
    layout["probe"].update(panel_probe())
    return layout


def main() -> int:
    threading.Thread(target=poll_metrics, daemon=True).start()
    threading.Thread(target=run_locust, daemon=True).start()
    with Live(render(), refresh_per_second=2, console=console, screen=True) as live:
        while state["phase"] != "DONE":
            time.sleep(0.5)
            live.update(render())
        live.update(render())
        time.sleep(3)
    console.print(panel_verdict())
    return 0 if state["verdict"] == "SHIP" else 1


if __name__ == "__main__":
    sys.exit(main())
