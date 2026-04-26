# agentsail

Dispatcher recipe. Use the terminal CLI directly:

```bash
agentsail init .
agentsail collect acme-bank
agentsail compile --customer acme-bank
agentsail check --customer acme-bank --target mock:support_agent_v12
agentsail verdict --customer acme-bank
agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --cmux-alert
agentsail report .agentsail/runs/acme-bank-run-001.json --open
agentsail loadtest run --config agentsail.loadtest.yaml --dry-run
```
