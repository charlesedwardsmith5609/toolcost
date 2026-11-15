# Step 4: cross-review scripts fail closed (new finding from cross-review)

## Behavior
- Both scripts: if gitleaks is missing or fails, or finds a secret, nothing is sent (exit 1).
- Only tracked changes are sent. The scripts no longer run `git add --intent-to-add --all`.
- Untracked files are sent only when named with `--include PATH` (`-Include` in the .ps1).
- A deny list (`.env*`, key and certificate files, `*credential*`, `.netrc`, `.npmrc`, `.pypirc`,
  `.aws/`, `.ssh/`, `.gnupg/`, terraform state and vars) is excluded from the diff and refused for
  `--include` (exit 2). Symlinks, directories and paths outside the repo are refused.
- `.sh` and `.ps1` share one deny list; a test compares the two lists.
- plan-review.sh shares `review_scan_secrets`, so it fails closed too.

## Ownership
No shared state. The test owns a temp git repo and a loopback HTTP server (127.0.0.1).

## Tests (scripts/crossreview_test.go, failing first against the old scripts)
- No gitleaks, gitleaks exit 7, gitleaks exit 3: no request reaches the server.
- Default run sends the tracked change, not untracked `.env`, untracked notes, or a modified tracked `server.key`; the index is unchanged.
- `--include notes.txt` sends it; `.env`, `.env.local`, `id_ed25519`, `*credentials*`, `.aws/`, tracked key, symlink and `../` paths are refused with exit 2.
- Deny lists in .sh and .ps1 are identical.
- The .ps1 could not be run here (no pwsh); only the list comparison covers it.

## Files
- scripts/cross-review.sh
- scripts/cross-review.ps1
- scripts/review-call.sh
- scripts/crossreview_test.go
