# Decisions

## 2026-11-13: Local repo, GitHub create not available

Context: `toolcost` was not on GitHub. `gh` is not logged in.

Choice: scaffold with agent-pairing-kit `new-project.sh` and develop on `improve/tldr-catchup-2026-11-13`. Push stays blocked. `go 1.26.11` matches the toolchain that passes govulncheck. The scaffold's 1.22.2 stdlib does not.

## 2026-11-13 phase 1 review

Finding: unpinned CI actions. Resolution: pin checkout and setup-go to the same commits used in iam-trust-audit. No second model was available for cross-review.sh.

## 2026-11-13 phase 2 review

Finding: encoding/json map order would change the count between runs. Resolution: canonical JSON sorts object keys before the estimator sees the bytes. Numbers stay as written via UseNumber. No second model was available.

## 2026-11-13 phase 3 review

Finding: a hard-coded divisor cannot be tested without depending on the byte length. Resolution: Count takes an Estimator. The CLI uses bytes/4. A unit test passes a stub that returns a fixed count. The error bound is documented as about 0.5x to 2x a typical BPE count, which is a claim about the method, not a measurement against a vendor tokenizer. No second model was available.

## 2026-11-13 phase 4 review

Finding: printing schemas would copy secrets that were pasted into a description. Resolution: text and JSON list names, kinds, counts, and suggestions. The CLI test asserts the schema key input_schema is absent from the JSON report. No second model was available.

## 2026-11-13 phase 5 review

Finding: a reader could treat the count as a vendor token bill. Resolution: the README says it is an estimate, states the divisor, and repeats the error bound. No second model was available.
