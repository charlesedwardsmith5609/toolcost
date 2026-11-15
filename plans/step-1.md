# Step 1: HIGH, input safety (findings 1, 2, 3)

## Types
- `load.ReadFile(path) ([]Item, error)`: open once, `Stat`, require regular, `io.LimitReader(f, MaxBytes+1)`. Open with `O_NOFOLLOW|O_NONBLOCK` (unix file; a fallback file for other OSes keeps the IsRegular check).
- `load.Parse`: decode with `Decoder`, then require the next `Token()` to be `io.EOF`. Error if the root object has neither `tools` nor `mcpServers`, or has unknown top-level keys. Depth cap 64 and duplicate-key rejection via a `Token()` walk before the `any` decode.
- Errors from load make `main` exit 2.

## Ownership
No shared state. Each call owns its buffers. `main` is single goroutine.

## Tests (written failing first)
- FIFO path returns an error within a 5 s timeout (ReadFile runs in a goroutine; the test selects on a timer).
- `/dev/zero` is refused as not regular.
- File of MaxBytes+1 bytes is refused; exactly MaxBytes is accepted.
- Symlink to a regular file is refused (kept).
- Table: `{}`, `{"foo":1}`, `null`, `{"tools":null}`, `{"tools":{}}` error; `{"tools":[]}` is accepted and reports 0.
- Table: `[] []`, `{"tools":[]} x`, `{"tools":[]}{"tools":[]}` error.
- Duplicate key and depth 65 nesting error.
- CLI: exit code is exactly 2 for each bad input (ExitError.ExitCode()).

## Files
- internal/load/load.go
- internal/load/open_unix.go
- internal/load/open_other.go
- internal/load/load_test.go
- cmd/toolcost/main_test.go
- README.md
