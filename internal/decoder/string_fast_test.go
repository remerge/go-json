package decoder

import (
	"bytes"
	"testing"
)

func scalarUnescapedStringEnd(buf []byte, cursor int64) (int64, bool) {
	for i := int(cursor); i < len(buf); i++ {
		switch buf[i] {
		case '"':
			return int64(i), true
		case '\\', 0:
			return int64(i), false
		}
	}
	return int64(len(buf)), false
}

func TestFastUnescapedStringEndBoundaries(t *testing.T) {
	for length := 0; length <= 40; length++ {
		for _, terminator := range []byte{'"', '\\', 0} {
			buf := append(bytes.Repeat([]byte{'a'}, length), terminator)
			buf = append(buf, bytes.Repeat([]byte{'z'}, 9)...)
			gotEnd, gotOK := fastUnescapedStringEnd(buf, 0)
			wantEnd, wantOK := scalarUnescapedStringEnd(buf, 0)
			if gotEnd != wantEnd || gotOK != wantOK {
				t.Fatalf("length %d terminator %q: got (%d, %v), want (%d, %v)", length, terminator, gotEnd, gotOK, wantEnd, wantOK)
			}
		}
	}
}

func TestFastUnescapedStringEndOffsetAndUTF8(t *testing.T) {
	buf := append([]byte("prefix:"), []byte("Grüße 東京 🙂\"")...)
	gotEnd, gotOK := fastUnescapedStringEnd(buf, int64(len("prefix:")))
	wantEnd, wantOK := scalarUnescapedStringEnd(buf, int64(len("prefix:")))
	if gotEnd != wantEnd || gotOK != wantOK {
		t.Fatalf("got (%d, %v), want (%d, %v)", gotEnd, gotOK, wantEnd, wantOK)
	}
}

func FuzzFastUnescapedStringEnd(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("short\""),
		[]byte("1234567\\tail"),
		[]byte("12345678\x00tail"),
		[]byte("Grüße 東京 🙂\""),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, buf []byte) {
		gotEnd, gotOK := fastUnescapedStringEnd(buf, 0)
		wantEnd, wantOK := scalarUnescapedStringEnd(buf, 0)
		if gotEnd != wantEnd || gotOK != wantOK {
			t.Fatalf("input %x: got (%d, %v), want (%d, %v)", buf, gotEnd, gotOK, wantEnd, wantOK)
		}
	})
}
