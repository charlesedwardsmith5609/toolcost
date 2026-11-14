# toolcost

Estimates the token cost of an agent tool setup. It reads JSON tool definitions and MCP server definitions, counts approximate tokens for each schema, and prints the total context overhead for one turn, the largest entries, and a suggestion to disable the large ones when the task does not need them.

The default estimator is `len(canonical JSON) / 4`, rounded up. Object keys are sorted before counting so the number does not depend on map order. This is not a vendor tokenizer and it does not use the network. For schema text the error bound is about 0.5x to 2x a typical BPE count. Pass another estimator to `estimate.Count` in tests or in a caller. The CLI uses the default.

```bash
go run ./cmd/toolcost --format text tools.json
go run ./cmd/toolcost --format json --budget 2000 tools.json
bash scripts/check.sh
```

Exit 0 when the total is within `--budget` or when no budget is set. Exit 1 when the total is greater than the budget. Exit 2 when a file cannot be read or parsed, or the budget is not a non-negative integer. The report does not print schema bodies.

Input is read with a single open. Symlinks and anything that is not a regular file (FIFOs, devices, directories) are refused, and files over 1 MiB (1048576 bytes) are refused; the read itself is bounded at that size. Input with no `tools` or `mcpServers`, unknown top-level keys, duplicate keys, nesting deeper than 64, or a second JSON value after the first is an error (exit 2).
