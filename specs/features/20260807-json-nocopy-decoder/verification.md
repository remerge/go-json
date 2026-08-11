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
- Clean integration from current fork `master`: passed normal, race, and focused fuzz validation.

## Contract Evidence

Explicit tests prove existing remerge interface-routing behavior is preserved for both APIs.
Explicit tests prove direct slice partial mutation is confined to `UnmarshalNoCopy`.
Reuse tests cover growing, shrinking, nested, pointer, and empty slices without stale decoded values.

## Performance Evidence

Final counterbalanced Darwin arm64 decoding improved 12.83% to 19.55% per fixture and 16.95% geometrically.
The Darwin decode-plus-native-parse pipeline improved 14.21% to 19.57% per fixture and 17.31% geometrically.
Decode allocation count fell 26.69% geometrically and pipeline allocation count fell 25.51%, while allocated bytes increased 0.62%.

An eight-sample final-head Linux amd64 refresh under OrbStack x86 emulation improved 14.96% to 19.62% per decode fixture and 18.05% geometrically.
The emulated Linux full pipeline improved 14.18% to 18.51% per fixture and 16.30% geometrically.
An earlier sixteen-sample emulated run on the pre-review candidate measured only 7.16% and 6.45% geometric improvements, respectively.
Both emulated runs reproduced the native arm64 allocation reductions.

The raw evidence and statistical summary are stored in the workspace Bet Execution Plan evidence directory.
The large difference between the two Linux runs demonstrates that timing under this emulator is directional only.
These measurements do not establish a production request distribution.

## Adoption Gates

The emulated Linux amd64 result is sufficient for this unmerged exploration.
Before any adoption, merge, or deployment decision, rerun the benchmark on production-like native Linux amd64 hardware and validate against a sanitized production-derived corpus.
