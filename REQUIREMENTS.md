# Requirements

`toolcost` estimates how many tokens a tool list and its MCP servers add to every model turn. It does not call a model and it does not use the network.

## Input

JSON. An object may contain `tools` (an array of tool objects) and `mcpServers` (a map of server definitions). A bare array is treated as `tools`.

## Estimator

The default estimator is `len(canonical JSON) / 4`, rounded up. Canonical JSON sorts object keys and keeps numbers as written. This is not a vendor tokenizer. The documented error bound is about 0.5x to 2x a typical BPE count for this kind of schema text. Callers can pass another estimator that implements the same `Tokens([]byte) int` method. No estimator may use the network.

## Output

Human text and JSON. The report has the total for one turn, the top offenders, and a short suggestion for each large entry. `--budget N` exits non-zero when the total is greater than N.

## Tests

Fixtures for a small setup under budget and a large setup over budget. A test uses a stub estimator so the count does not depend on the default divisor. Tests do not use the network.
