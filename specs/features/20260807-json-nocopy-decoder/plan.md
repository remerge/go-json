# No-Copy Decoder Implementation Plan

1. Synchronize with current upstream and retain remerge's custom API commits.
2. Add a portable word-at-a-time fast path for unescaped strings.
3. Preserve the existing remerge interface-routing behavior.
4. Mark only `UnmarshalNoCopy` runtime contexts with the no-copy option.
5. Dispatch slices to direct destination decoding only when the no-copy option is present.
6. Add contract, differential, malformed-input, reuse, concurrency, and fuzz tests.
7. Run cross-platform bidder benchmarks and document evidence.
8. Run normal, race, fuzz, and downstream bidder validation without merge or deployment.
