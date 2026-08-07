# No-Copy Decoder Verification

## Current Proof

- `go test ./...`: passed.
- `go test -race ./...`: passed.
- String-scanner fuzzing: passed for about 821,000 executions over eleven seconds.
- Slice differential fuzzing: passed for about 259,000 executions over sixteen seconds.
- Bidder benchmark compatibility and malformed-input tests: passed.
- Bidder `integration/ssp`: passed.
- Bidder `integration/ssp` under the race detector: passed.
- Linux amd64 container `go test ./...`: passed.
- Linux amd64 container `go test -race ./...`: passed on the final candidate.

## Contract Evidence

Explicit tests prove existing ReMerge interface-routing behavior is preserved for both APIs.
Explicit tests prove direct slice partial mutation is confined to `UnmarshalNoCopy`.
Reuse tests cover growing, shrinking, nested, pointer, and empty slices without stale decoded values.

## Performance Evidence

Final counterbalanced Darwin arm64 decoding improved 14.59% to 18.03% per fixture and 16.37% geometrically.
The Darwin decode-plus-native-parse pipeline improved 13.74% to 21.34% per fixture and 17.80% geometrically.
Decode allocation count fell 26.69% geometrically, while allocated bytes increased 0.62%.

Linux amd64 under OrbStack x86 emulation improved 3.37% to 10.55% per decode fixture and 7.16% geometrically.
The emulated Linux full pipeline improved 4.17% to 9.59% per fixture and 6.45% geometrically.
The emulated allocation results matched native arm64.

The raw evidence and statistical summary are stored in the workspace Bet Execution Plan evidence directory.
The Linux timing is directional only because it ran under emulation.
These measurements do not establish a production request distribution.

## Completion And Adoption Gates

- Local no-mistakes validation.

The emulated Linux amd64 result is sufficient for this unmerged exploration.
Before any adoption, merge, or deployment decision, rerun the benchmark on production-like native Linux amd64 hardware and validate against a sanitized production-derived corpus.
