# Threat model

The tool reads JSON the operator names. It does not send schemas to a tokenizer service. A schema can still contain a secret that was pasted into a description. The report prints tool names, byte-derived counts, and suggestions. It does not print the schema body.

## What the number is

The count is an estimate of context overhead per turn, not a bill and not a vendor's token counter. The error bound is about 0.5x to 2x a typical BPE count for schema text. Treat a number under the budget as a hint, and keep a margin.

## Failure

Symlinks, files over 1 MiB, and invalid JSON exit 2. A missing budget is not an error. A budget below zero is an error.
