# No-Copy Decoder Contracts

## Public API Boundary

`Unmarshal` retains the existing remerge decoder compiler behavior.
It bypasses `json.Unmarshaler` and `encoding.TextUnmarshaler` and prefers `RemergeUnmarshalJSON` where implemented.

`UnmarshalNoCopy` uses the same established interface routing and marks the runtime context with the internal no-copy option.

## Slice Mutation Contract

The ordinary slice decoder retains the current fork's staged decoding and assigns elements on successful array completion.
The no-copy slice decoder writes directly into destination capacity and grows that capacity when required.
The no-copy decoder clears each reused element before decoding into it.
After a no-copy error, the destination is invalid and may contain a partial result.

## String Scanning Contract

The fast scanner reads only complete eight-byte windows from the supplied slice.
It interprets each window with explicit little-endian byte order so byte position calculation is independent of host byte order.
It stops at the first quote, backslash, or NUL byte.
It returns to the existing scalar escape and error path when a backslash or NUL is encountered.

## Bidder Integration Contract

`UnmarshalNoCopy` accepts ordinary byte slices. When the caller already appends a trailing NUL sentinel, the decoder can preserve aliasing to that input and avoid the extra copy needed to add the sentinel itself.
When the caller does not supply the sentinel, `UnmarshalNoCopy` copies once to append it before decoding.
When the caller does supply the sentinel, it is accepted only as the final byte after any trailing JSON whitespace, and junk after a sentinel is rejected immediately.

The bidder continues to append a trailing NUL sentinel before calling `UnmarshalNoCopy`.
The bidder must discard the request object when no-copy decoding returns an error.
Strings decoded without escapes may alias the request input buffer only when the caller supplied the sentinel, and those strings must not outlive that buffer.
Escaped strings on that sentinel-backed path may also be unescaped by mutating the caller-owned buffer in place.
