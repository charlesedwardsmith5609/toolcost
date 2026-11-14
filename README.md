# toolcost

Estimates the token cost of an agent tool setup. It reads JSON tool definitions and MCP server definitions, counts approximate tokens for each entry, and prints the total context overhead for one turn, the largest entries, and suggestions to disable the large ones when the task does not need them. It does not use the network.

## Run

Requires Go (see `go.mod` for the version). The module uses only the standard library.

```bash
go run ./cmd/toolcost --format text tools.json
go run ./cmd/toolcost --format json --budget 2000 tools.json
bash scripts/check.sh
```

`bash scripts/check.sh` is the validation gate (gofmt, go vet, govulncheck, `go test -race -count=1`, and helper checks). It prints `CHECK PASSED` when every stage passes. Sample inputs are in `testdata/`.

## Flags

```
toolcost [--format text|json] [--budget N] file...
```

- `--format`: `text` (default) or `json`. Any other value is an error.
- `--budget N`: N is a non-negative integer; 0 is accepted. The text and JSON reports show the budget when it is set, including 0.
- At least one file is required. Several files are combined into one report.
- There is no flag for the estimator or its divisor. The CLI always uses the default.

## Input

JSON. Either an object with `tools` (an array) and/or `mcpServers` (an object keyed by server name), or a bare array, which is read as `tools`. A tool without a non-empty string `name` is shown as `tool[i]`, where i is its index.

## Estimator

The default estimator is `len(canonical JSON) / 4`, rounded up. Object keys are sorted and numbers are kept as written, so the count does not depend on map order. This is not a vendor tokenizer. The error against a vendor tokenizer is unmeasured, so treat the number as a rough estimate and keep a margin.

The report names the estimator: `bytes/N` for the built-in `BytesPerToken{Divisor: N}` (the CLI prints `bytes/4`), or `custom` for any other estimator passed to `estimate.Count`. `estimate.Count` rejects a `BytesPerToken` divisor outside 1..1048576 (2^20) instead of replacing it. This applies to library callers and tests only.

## Output

Text shows the total, the estimator, the budget (if set, plus `over budget by N` when exceeded), the MCP config total (if any), the top 5 rows, and suggestions. JSON has `total`, `mcp_config_tokens`, `budget` (present only when set), `estimator`, `rows` and `suggestions`. JSON is not HTML-escaped. Schema bodies are never printed.

Names are printed with Go `%q` quoting in text rows and in suggestions, and suggestions also escape control characters, so a name cannot inject terminal escapes or extra lines. JSON uses normal JSON string escaping.

MCP server entries measure launch config only. They are labeled "config only", reported as `mcp_config_tokens`, and left out of the total, of `--budget`, and of suggestions. `tools/list` is not read.

Suggestions ("Disable ... on turns that do not need it") appear only when the total is at least 200 tokens. A tool row gets one if it has at least 200 tokens, or if there are at least 3 tool rows and the row has at least 25% of the total (integer division).

## Exit codes

- 0: the total is within `--budget`, or no budget is set.
- 1: the total is greater than the budget. Only tool rows count.
- 2: usage error, bad `--budget` or `--format`, a file that cannot be read, or input that is rejected.

## Input safety

Each file is opened once, read-only, with `O_NOFOLLOW` and `O_NONBLOCK` on Unix. On other platforms it checks `Lstat` first, so a swap between the check and the open is possible there. The opened file must be a regular file: symlinks, FIFOs, devices and directories are refused. Files over 1 MiB (1048576 bytes) are refused, and the read is bounded at that size.

Input is rejected (exit 2) when it has no `tools` or `mcpServers`, has an unknown top-level key, has wrong types for `tools` or `mcpServers`, has a duplicate key at any depth, nests containers deeper than 64, has a second JSON value after the first, or is not valid JSON. It is never reported as 0 tokens.

## Cross-model review scripts

`scripts/cross-review.sh` (bash) and `scripts/cross-review.ps1` (PowerShell) send a diff to another AI provider for review. They are separate from the CLI and need provider credentials (see the header of each script).

- Only tracked changes are sent. Untracked files are sent only when named with `--include PATH` (`-Include` in PowerShell). With no changes and no `--base`/`-Base`, the last commit is reviewed. The diff is cut at 60000 characters.
- Paths matching a deny list are never sent, even when named: `.env*`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `*.jks`, `*.keystore`, `*.kdbx`, `id_rsa*`, `id_dsa*`, `id_ecdsa*`, `id_ed25519*`, `*credential*`, `.netrc`, `.npmrc`, `.pypirc`, `.git-credentials`, `*.tfstate`, `*.tfvars`, and anything under `.aws/`, `.ssh/` or `.gnupg/`. `go.sum` and `package-lock.json` are also left out. `--include` also refuses symlinks, non-files and paths outside the repository.
- The request is scanned with `gitleaks` first. If gitleaks is missing, fails, or finds a possible secret, nothing is sent (exit 1).
- The shell script supports `--provider cursor|openai|anthropic|auto`. The PowerShell script supports `-Provider openai|anthropic`.
- Reviews are saved under `.reviews/`, which git ignores.

`scripts/crossreview_test.go` tests the shell script against a local fake server and checks that both scripts carry the same deny list. The PowerShell script parses without errors and was run under PowerShell 7.7.0 for the fail-closed paths (missing gitleaks, denied `-Include`); the full send path through the `.ps1` is not exercised by the tests.
