package json_test

import (
	stdjson "encoding/json"
	"reflect"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/goccy/go-json"
)

type noCopySliceContainer struct {
	Values []int `json:"values"`
}

type dualUnmarshaler struct {
	Decoder string
}

func (v *dualUnmarshaler) UnmarshalJSON([]byte) error {
	v.Decoder = "standard"
	return nil
}

func (v *dualUnmarshaler) RemergeUnmarshalJSON([]byte) error {
	v.Decoder = "remerge"
	return nil
}

type standardOnlyUnmarshaler struct {
	Value  string `json:"value"`
	Called bool   `json:"-"`
}

type differentialSliceContainer struct {
	Values  []int                 `json:"values"`
	Nested  [][]string            `json:"nested"`
	Objects []*differentialObject `json:"objects"`
	Flags   []bool                `json:"flags"`
}

type differentialObject struct {
	Name string `json:"name"`
}

type concurrentNoCopyContainer struct {
	Items []struct {
		ID   string `json:"id"`
		Tags []int  `json:"tags"`
	} `json:"items"`
}

func (v *standardOnlyUnmarshaler) UnmarshalJSON(data []byte) error {
	type plain standardOnlyUnmarshaler
	if err := stdjson.Unmarshal(data, (*plain)(v)); err != nil {
		return err
	}
	v.Called = true
	return nil
}

func TestUnmarshalNoCopyScopesDirectSliceMutation(t *testing.T) {
	invalid := []byte(`{"values":[1,2,]}`)

	t.Run("ordinary unmarshal preserves destination on array syntax error", func(t *testing.T) {
		got := noCopySliceContainer{Values: []int{9, 8}}
		if err := json.Unmarshal(invalid, &got); err == nil {
			t.Fatal("expected syntax error")
		}
		want := noCopySliceContainer{Values: []int{9, 8}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ordinary Unmarshal mutated destination: got %#v, want %#v", got, want)
		}
	})

	t.Run("no-copy unmarshal may expose partial destination on error", func(t *testing.T) {
		input := append(append([]byte(nil), invalid...), 0)
		got := noCopySliceContainer{Values: []int{9, 8}}
		if err := json.UnmarshalNoCopy(input, &got); err == nil {
			t.Fatal("expected syntax error")
		}
		want := noCopySliceContainer{Values: []int{1, 2}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("no-copy partial result: got %#v, want %#v", got, want)
		}
	})
}

func TestUnmarshalNoCopyMatchesUnmarshalForValidSlices(t *testing.T) {
	input := []byte(`{"values":[1,2,3,4]}`)
	var ordinary noCopySliceContainer
	if err := json.Unmarshal(input, &ordinary); err != nil {
		t.Fatalf("ordinary Unmarshal: %v", err)
	}

	input = append(input, 0)
	var noCopy noCopySliceContainer
	if err := json.UnmarshalNoCopy(input, &noCopy); err != nil {
		t.Fatalf("UnmarshalNoCopy: %v", err)
	}
	if !reflect.DeepEqual(noCopy, ordinary) {
		t.Fatalf("result mismatch: no-copy %#v, ordinary %#v", noCopy, ordinary)
	}
}

func TestUnmarshalNoCopyUsesDedicatedInterfaceSemantics(t *testing.T) {
	t.Run("remerge interface is limited to no-copy", func(t *testing.T) {
		var ordinary dualUnmarshaler
		if err := json.Unmarshal([]byte(`{}`), &ordinary); err != nil {
			t.Fatalf("ordinary Unmarshal: %v", err)
		}
		if ordinary.Decoder != "standard" {
			t.Fatalf("ordinary decoder = %q, want standard", ordinary.Decoder)
		}

		var noCopy dualUnmarshaler
		if err := json.UnmarshalNoCopy([]byte{'{', '}', 0}, &noCopy); err != nil {
			t.Fatalf("UnmarshalNoCopy: %v", err)
		}
		if noCopy.Decoder != "remerge" {
			t.Fatalf("no-copy decoder = %q, want remerge", noCopy.Decoder)
		}
	})

	t.Run("standard interface is bypassed only by no-copy", func(t *testing.T) {
		input := []byte(`{"value":"decoded"}`)
		var ordinary standardOnlyUnmarshaler
		if err := json.Unmarshal(input, &ordinary); err != nil {
			t.Fatalf("ordinary Unmarshal: %v", err)
		}
		if !ordinary.Called || ordinary.Value != "decoded" {
			t.Fatalf("ordinary result = %#v", ordinary)
		}

		input = append(input, 0)
		var noCopy standardOnlyUnmarshaler
		if err := json.UnmarshalNoCopy(input, &noCopy); err != nil {
			t.Fatalf("UnmarshalNoCopy: %v", err)
		}
		if noCopy.Called || noCopy.Value != "decoded" {
			t.Fatalf("no-copy result = %#v", noCopy)
		}
	})
}

func TestUnmarshalNoCopyReusesSliceCapacityWithoutStaleValues(t *testing.T) {
	inputs := [][]byte{
		[]byte(`{"values":[1,2,3,4],"nested":[["a","b"],["c"]],"objects":[{"name":"first"},{"name":"second"}],"flags":[true,false]}`),
		[]byte(`{"values":[9],"nested":[["short"]],"objects":[{"name":"only"}],"flags":[]}`),
		[]byte(`{"values":[5,6,7],"nested":[],"objects":[null,{"name":"last"}],"flags":[false,true,true]}`),
	}

	var got differentialSliceContainer
	for _, input := range inputs {
		var want differentialSliceContainer
		if err := json.Unmarshal(input, &want); err != nil {
			t.Fatalf("ordinary Unmarshal: %v", err)
		}
		noCopyInput := append(append([]byte(nil), input...), 0)
		if err := json.UnmarshalNoCopy(noCopyInput, &got); err != nil {
			t.Fatalf("UnmarshalNoCopy: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("result mismatch after reuse: no-copy %#v, ordinary %#v", got, want)
		}
	}
}

func TestConcurrentStandardAndNoCopyDecoderCompilation(t *testing.T) {
	const workers = 64
	input := []byte(`{"items":[{"id":"a","tags":[1,2,3]},{"id":"b","tags":[4]}]}`)
	start := make(chan struct{})
	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(noCopy bool) {
			defer wg.Done()
			<-start
			var got concurrentNoCopyContainer
			if noCopy {
				data := append(append([]byte(nil), input...), 0)
				errCh <- json.UnmarshalNoCopy(data, &got)
				return
			}
			errCh <- json.Unmarshal(input, &got)
		}(i%2 == 0)
	}
	close(start)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func FuzzUnmarshalNoCopySlicesMatchStandard(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{}`),
		[]byte(`{"values":[]}`),
		[]byte(`{"values":[1,2,3],"nested":[["a"],["b","c"]]}`),
		[]byte(`{"objects":[null,{"name":"escaped\\nvalue"}],"flags":[true,false]}`),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		if !utf8.Valid(input) || !stdjson.Valid(input) {
			return
		}

		var stdlib differentialSliceContainer
		stdlibErr := stdjson.Unmarshal(input, &stdlib)
		var ordinary differentialSliceContainer
		ordinaryErr := json.Unmarshal(input, &ordinary)
		noCopyInput := append(append([]byte(nil), input...), 0)
		var noCopy differentialSliceContainer
		noCopyErr := json.UnmarshalNoCopy(noCopyInput, &noCopy)

		if (stdlibErr == nil) != (ordinaryErr == nil) || (ordinaryErr == nil) != (noCopyErr == nil) {
			t.Fatalf("success mismatch: stdlib=%v ordinary=%v no-copy=%v input=%s", stdlibErr, ordinaryErr, noCopyErr, input)
		}
		if stdlibErr != nil {
			return
		}
		if !reflect.DeepEqual(ordinary, stdlib) {
			t.Fatalf("ordinary mismatch: got %#v, stdlib %#v", ordinary, stdlib)
		}
		if !reflect.DeepEqual(noCopy, ordinary) {
			t.Fatalf("no-copy mismatch: got %#v, ordinary %#v", noCopy, ordinary)
		}
	})
}
