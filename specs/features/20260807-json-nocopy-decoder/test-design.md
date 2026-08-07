# No-Copy Decoder Test Design

## Functional Coverage

- Run the fork test suite and explicitly classify inherited upstream conformance differences.
- Verify ordinary and no-copy interface routing remains consistent with the current ReMerge fork for types implementing standard, ReMerge, or both interfaces.
- Verify valid direct-slice decoding across empty, growing, shrinking, nested, pointer, and reused slices.
- Verify ordinary malformed-array input does not expose direct partial mutation.
- Verify no-copy malformed-array input documents its permitted partial result.

## String Scanner Coverage

- Test quote, backslash, and NUL at every offset around multiple eight-byte boundaries.
- Test non-zero starting offsets and multibyte UTF-8.
- Differentially fuzz the fast scanner against a scalar reference.

## Differential And Fuzz Coverage

- For valid UTF-8 JSON, compare standard library, ordinary fork, and no-copy fork success and decoded values for plain slice-bearing structures.
- Fuzz the comparison for at least ten seconds during local validation.
- Preserve interesting generated corpus entries in the Go fuzz cache rather than committing machine-generated corpus files.

## Concurrency Coverage

- Exercise ordinary and no-copy decoding for the same new type concurrently.
- Run the full suite with the race detector.

## Bidder Coverage

- Run the six-fixture compatibility and malformed-input benchmark tests.
- Run bidder `integration/ssp` normally and with the race detector.
- Benchmark decode-only and native parse-pipeline paths against the current fork.

## Platform Coverage

- Measure Darwin arm64 for local development evidence.
- Measure Linux amd64 before recommending adoption.
