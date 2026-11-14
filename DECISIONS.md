# Decisions

## 2026-11-13: Local repo, GitHub create not available

Context: `toolcost` was not on GitHub. `gh` is not logged in.

Choice: scaffold with agent-pairing-kit `new-project.sh` and develop on `improve/tldr-catchup-2026-11-13`. Push stays blocked. `go 1.26.11` matches the toolchain that passes govulncheck. The scaffold's 1.22.2 stdlib does not.

## 2026-11-13 phase 1 review

Finding: unpinned CI actions. Resolution: pin checkout and setup-go to the same commits used in iam-trust-audit. No second model was available for cross-review.sh.
