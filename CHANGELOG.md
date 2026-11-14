# Changelog

## 2026-11-14

Changes on `improve/tldr-catchup-2026-11-13` relative to `origin/main`.

### Changed

- Input reading: one open per file, with `O_NOFOLLOW` and `O_NONBLOCK` on Unix; regular files only; 1 MiB bounded read. Symlinks, FIFOs, devices and directories exit 2.
- Input validation: unknown top-level keys, a missing `tools`/`mcpServers`, wrong types, duplicate keys, nesting deeper than 64 and trailing JSON values are errors (exit 2). They no longer report 0 tokens.
- Estimator: the report names the estimator (`bytes/N` or `custom`). A divisor outside 1..2^20 is rejected rather than replaced. The "0.5x to 2x" error bound was removed from the README and threat model because it was never measured.
- MCP server rows are labeled "config only", summed as `mcp_config_tokens`, and excluded from the total, `--budget` and suggestions.
- Suggestions: none when the total is under 200 tokens; a tool row gets one at 200 tokens or more, or at 25% or more of the total when there are at least 3 tool rows.
- Report: the budget (including 0) is shown in text and JSON; `over budget by N` is printed when exceeded; JSON output is not HTML-escaped; names are printed with `%q` and suggestions escape control characters.
- `scripts/cross-review.sh` and `.ps1`: fail closed (no gitleaks, scan failure or finding sends nothing); send tracked changes only; `--include`/`-Include` names untracked files; a secret-file deny list is applied; `.ps1` parse error fixed (commit 9f6ae0b) and checked under PowerShell 7.7.0.
- `scripts/check.sh`, `assertion-check.sh`, `misbehavior-check.sh`, `review-call.sh`: refreshed from agent-pairing-kit.

### Added

- `scripts/crossreview_test.go`, plus tests for the load, estimate, report and CLI changes (exact exit codes, table tests).
- `plans/step-1.md` to `plans/step-4.md` and the 2026-11-14 review fixes in `REQUIREMENTS.md`.
- This file.

### Docs

- README, `docs/threat-model.md`, `DECISIONS.md`, `CLAUDE.md` and `AGENTS.md` updated to match the code.
