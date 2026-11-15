# Requirements

`toolcost` estimates how many tokens a tool list and its MCP servers add to every model turn. It does not call a model and it does not use the network.

## Input

JSON. An object may contain `tools` (an array of tool objects) and `mcpServers` (a map of server definitions). A bare array is treated as `tools`.

## Estimator

The default estimator is `len(canonical JSON) / 4`, rounded up. Canonical JSON sorts object keys and keeps numbers as written. This is not a vendor tokenizer. The error against a vendor tokenizer is unmeasured. Callers can pass another estimator that implements the same `Tokens([]byte) int` method. No estimator may use the network.

## Output

Human text and JSON. The report has the total for one turn, the top offenders, and a short suggestion for each large entry. `--budget N` exits non-zero when the total is greater than N.

## Tests

Fixtures for a small setup under budget and a large setup over budget. A test uses a stub estimator so the count does not depend on the default divisor. Tests do not use the network.

## Review fixes 2026-11-14

- Input files are read with one open, a regular-file check, and a 1 MiB bounded read. FIFOs, devices and symlinks are refused with exit 2.
- Input with an unrecognized shape (no `tools` and no `mcpServers`, wrong types, unknown top-level keys) is an error with exit 2. It never reports 0 tokens.
- Trailing JSON values after the first document are an error with exit 2.
- MCP server numbers measure launch config only. They are labeled "config only" and are excluded from the total and from `--budget`. Reading `tools/list` is out of scope.
- The report prints only what is measured: the estimator name and divisor. The 0.5x to 2x bound is not printed or documented as measured.
- A divisor that would overflow or is non-positive is rejected, not silently replaced.
- Tool and server names are printed quoted so control characters cannot reach the terminal.
- The suggestion rule depends on setup size so small setups get no suggestion.
- JSON output does not HTML-escape. The text and JSON report show the budget, including 0.
- CLI tests assert exact exit codes (1 for over budget, 2 for errors) and use table tests.
