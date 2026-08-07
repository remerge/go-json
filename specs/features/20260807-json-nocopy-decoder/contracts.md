# No-Copy Decoder Contracts

## Public API Boundary

`Unmarshal` retains the existing ReMerge decoder compiler behavior.
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

The bidder continues to append a trailing NUL sentinel before calling `UnmarshalNoCopy`.
The bidder must discard the request object when no-copy decoding returns an error.
Strings decoded without escapes may alias the request input buffer and must not outlive that buffer.
