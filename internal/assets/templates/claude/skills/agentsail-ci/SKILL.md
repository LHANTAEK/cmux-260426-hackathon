# Agent Sail CI

Run compile, check, verdict, and report sequentially.

## Workflow

1. Parse `--customer <customer>` and `--target <url>`.
2. If `.agentsail/cache/<customer>/` is empty, stop and ask for `/agentsail collect <customer>`.
3. Prefer the Go CLI:
   - `bin/agentsail ci --customer <customer> --target <url> [--report] [--open]`
   - fallback: `go run ./cmd/agentsail ci --customer <customer> --target <url> [--report] [--open]`
4. Emit cmux OSC9 alerts for phase changes when the CLI does not.

## Expected Demo Verdicts

- `finbank`: `HOLD`
- `retailco`: `SHIP`
- `acme-bank`: `BLOCK`
