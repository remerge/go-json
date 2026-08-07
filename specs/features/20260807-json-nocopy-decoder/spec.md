# No-Copy Decoder Update

## Purpose

Update the ReMerge fork to current upstream hardening while preserving its bidder-specific no-copy decoding behavior.
Improve OpenRTB decoding through portable word-at-a-time string scanning and direct destination slice decoding.

## Required Behavior

- `Unmarshal` must retain the existing ReMerge fork's interface-routing semantics.
- `UnmarshalNoCopy` must retain ReMerge's direct-buffer and `RemergeUnmarshalJSON` behavior.
- Standard JSON and text unmarshal interfaces must remain bypassed as they are in the existing ReMerge fork.
- Direct destination slice decoding must run only on the no-copy path.
- Valid no-copy results must match ordinary and standard-library results for supported plain Go values.
- The no-copy caller continues to own input-buffer lifetime and must provide the existing trailing NUL sentinel.

## Error Behavior

Ordinary `Unmarshal` must retain upstream behavior when malformed array input is encountered.
`UnmarshalNoCopy` may expose a partially decoded destination after an error because it writes array elements directly into destination storage.
Callers must discard the destination after a no-copy decoding error.

## Non-Goals

- This feature does not add Sonic JIT or SIMD assembly.
- This feature does not change bidder request handling.
- This feature does not authorize merge, release, deployment, or traffic changes.
- This feature does not claim production-wide performance from the checked-in fixture corpus.
