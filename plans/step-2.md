# Step 2: MEDIUM, honest numbers (findings 4, 5, 6 and DECISIONS D1)

## Types
- `estimate.BytesPerToken`: add `Validate() error` (Divisor between 1 and 1<<20); `Tokens` no longer silently replaces a bad divisor. `Default()` stays 4. Sums use an overflow-checked add.
- Optional `estimate.Named` interface (`Name() string`). `Report.Estimator` comes from it (`bytes/4`, `bytes/8`, `custom`). The `ErrorBound` field is removed.
- `Report` gains `MCPConfigTokens int` (`json:"mcp_config_tokens"`). `Total` and the budget comparison count `tool` rows only. `Row` gains `Note` set to "config only" for `mcp`.
- `report.Text` prints names with `%q` and labels mcp rows "config only".

## Ownership
No shared state; `Count` builds a fresh Report per call.

## Tests (failing first)
- Divisor 0, negative and 1<<62 rejected by Validate; a huge sum does not go negative.
- Estimator name follows the divisor (`bytes/8`); a stub reports `custom`.
- Text and JSON contain no "0.5x" bound; README wording changed to "unmeasured".
- Name `"evil\x1b[31m"` prints escaped with no raw ESC byte.
- MCP-only file: total 0, mcp_config_tokens > 0, `--budget 0` exits 0. Tools plus MCP: only tools count toward the budget.
- CLI exit codes exactly 0, 1, 2.

## Files
- internal/estimate/estimate.go
- internal/estimate/estimate_test.go
- internal/report/report.go
- internal/report/report_test.go
- cmd/toolcost/main.go
- cmd/toolcost/main_test.go
- README.md
- REQUIREMENTS.md
- docs/threat-model.md
