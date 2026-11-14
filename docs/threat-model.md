# Threat model

The tool reads JSON the operator names. It does not send schemas to a tokenizer service. A schema can still contain a secret that was pasted into a description. The report prints tool names, byte-derived counts, and suggestions. It does not print the schema body.

## What the number is

The count is an estimate of context overhead per turn, not a bill and not a vendor's token counter. The error against a vendor tokenizer is unmeasured. Treat a number under the budget as a hint, and keep a margin.

## Failure

Exit 2 for: symlinks and non-regular files (FIFOs, devices, directories), files over 1 MiB, invalid JSON, unknown input shape or top-level keys, duplicate keys, nesting deeper than 64, and a second JSON value after the first. A missing budget is not an error. A budget below zero or a non-integer budget is an error. Exit 1 means the total is over the budget.

## Names

Tool and server names come from the input. Text output quotes them with `%q` and escapes control characters in suggestions, so a name cannot forge lines or terminal escapes. JSON output uses JSON escaping and is not HTML-escaped.

## MCP servers

MCP rows count the whole server definition from the input (its launch config; `tools/list` is not read). They are reported separately as `mcp_config_tokens`. The report does not print their bodies, but the input file may contain secrets in `env`, so do not send it to a third party.

## Cross-model review scripts

`scripts/cross-review.sh` and `.ps1` send a diff to an external provider. They send tracked changes only, apply a secret-file deny list, and scan the request with gitleaks first. A missing or failing gitleaks, or a finding, means nothing is sent. Untracked files are sent only when named with `--include` or `-Include`. gitleaks can miss a secret, so read the diff before running a review.

