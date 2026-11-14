# Step 3: LOW, output polish and tests (findings 7, 8, 9)

## Types
- `estimate.Count`: no suggestions when `Total < 200`. Otherwise a row qualifies at `Tokens >= 200`, or at share >= 25% when there are at least 3 rows.
- `report.JSON`: `enc.SetEscapeHTML(false)`.
- `Report.Budget` becomes `*int` (`json:"budget,omitempty"`) so budget 0 is shown. Text prints `budget N` and `over budget by M` when exceeded.

## Ownership
No shared state.

## Tests (failing first)
- Table: one tool at 100% share and 50 tokens gets no suggestion; 3 rows with one at 30% and Total 500 gets one; Total 199 never does.
- JSON output with name `a<b>&c` contains the literal characters, not `<`.
- `--budget 0` prints and exits per rule; JSON includes `"budget": 0`.
- Table test over fixtures asserting exact `ExitError.ExitCode()` values (0, 1, 2), replacing the `err == nil` check.

## Files
- internal/estimate/estimate.go
- internal/estimate/estimate_test.go
- internal/report/report.go
- internal/report/report_test.go
- cmd/toolcost/main.go
- cmd/toolcost/main_test.go
