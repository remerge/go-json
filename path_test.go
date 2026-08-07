package json_test

import (
	"bytes"
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/goccy/go-json"
)

func TestExtractPath(t *testing.T) {
	src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
	t.Run("$.a.b", func(t *testing.T) {
		path, err := json.CreatePath("$.a.b")
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 1 {
			t.Fatal("failed to extract")
		}
		if !bytes.Equal(contents[0], []byte("10")) {
			t.Fatal("failed to extract")
		}
	})
	t.Run("$.b", func(t *testing.T) {
		path, err := json.CreatePath("$.b")
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 1 {
			t.Fatal("failed to extract")
		}
		if !bytes.Equal(contents[0], []byte(`"text"`)) {
			t.Fatal("failed to extract")
		}
	})
	t.Run("$.a", func(t *testing.T) {
		path, err := json.CreatePath("$.a")
		if err != nil {
			t.Fatal(err)
		}
		contents, err := path.Extract(src)
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) != 1 {
			t.Fatal("failed to extract")
		}
		if !bytes.Equal(contents[0], []byte(`{"b":10,"c":true}`)) {
			t.Fatal("failed to extract")
		}
	})
}

func TestUnmarshalPath(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
		t.Run("success", func(t *testing.T) {
			path, err := json.CreatePath("$.a.b")
			if err != nil {
				t.Fatal(err)
			}
			var v int
			if err := path.Unmarshal(src, &v); err != nil {
				t.Fatal(err)
			}
			if v != 10 {
				t.Fatal("failed to unmarshal path")
			}
		})
		t.Run("failure", func(t *testing.T) {
			path, err := json.CreatePath("$.a.c")
			if err != nil {
				t.Fatal(err)
			}
			var v map[string]interface{}
			if err := path.Unmarshal(src, &v); err == nil {
				t.Fatal("expected error")
			}
		})
	})
	t.Run("bool", func(t *testing.T) {
		src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
		t.Run("success", func(t *testing.T) {
			path, err := json.CreatePath("$.a.c")
			if err != nil {
				t.Fatal(err)
			}
			var v bool
			if err := path.Unmarshal(src, &v); err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Fatal("failed to unmarshal path")
			}
		})
		t.Run("failure", func(t *testing.T) {
			path, err := json.CreatePath("$.a.b")
			if err != nil {
				t.Fatal(err)
			}
			var v bool
			if err := path.Unmarshal(src, &v); err == nil {
				t.Fatal("expected error")
			}
		})
	})
	t.Run("map", func(t *testing.T) {
		src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
		t.Run("success", func(t *testing.T) {
			path, err := json.CreatePath("$.a")
			if err != nil {
				t.Fatal(err)
			}
			var v map[string]interface{}
			if err := path.Unmarshal(src, &v); err != nil {
				t.Fatal(err)
			}
			if len(v) != 2 {
				t.Fatal("failed to decode map")
			}
		})
	})

	t.Run("path with single quote selector", func(t *testing.T) {
		path, err := json.CreatePath("$['a.b'].c")
		if err != nil {
			t.Fatal(err)
		}

		var v string
		if err := path.Unmarshal([]byte(`{"a.b": {"c": "world"}}`), &v); err != nil {
			t.Fatal(err)
		}
		if v != "world" {
			t.Fatal("failed to unmarshal path")
		}
	})
	t.Run("path with double quote selector", func(t *testing.T) {
		path, err := json.CreatePath(`$."a.b".c`)
		if err != nil {
			t.Fatal(err)
		}

		var v string
		if err := path.Unmarshal([]byte(`{"a.b": {"c": "world"}}`), &v); err != nil {
			t.Fatal(err)
		}
		if v != "world" {
			t.Fatal("failed to unmarshal path")
		}
	})
	t.Run("rejects out of range integers", func(t *testing.T) {
		t.Run("negative into uint", func(t *testing.T) {
			path, err := json.CreatePath("$.value")
			if err != nil {
				t.Fatal(err)
			}
			v := uint(7)
			if err := path.Unmarshal([]byte(`{"value":-1}`), &v); err == nil {
				t.Fatal("expected error")
			}
			if v != 7 {
				t.Fatalf("destination mutated: got %d, want %d", v, 7)
			}
		})
		t.Run("width overflow into int8", func(t *testing.T) {
			path, err := json.CreatePath("$.value")
			if err != nil {
				t.Fatal(err)
			}
			v := int8(7)
			if err := path.Unmarshal([]byte(`{"value":300}`), &v); err == nil {
				t.Fatal("expected error")
			}
			if v != 7 {
				t.Fatalf("destination mutated: got %d, want %d", v, 7)
			}
		})
		t.Run("fractional float into int", func(t *testing.T) {
			path, err := json.CreatePath("$.value")
			if err != nil {
				t.Fatal(err)
			}
			v := 7
			if err := path.Unmarshal([]byte(`{"value":1.5}`), &v); err == nil {
				t.Fatal("expected error")
			}
			if v != 7 {
				t.Fatalf("destination mutated: got %d, want %d", v, 7)
			}
		})
		t.Run("imprecise float-backed integer into int64", func(t *testing.T) {
			path, err := json.CreatePath("$.value")
			if err != nil {
				t.Fatal(err)
			}
			v := int64(7)
			if err := path.Unmarshal([]byte(`{"value":9007199254740993}`), &v); err != nil {
				t.Fatal(err)
			}
			if v != 9007199254740993 {
				t.Fatalf("destination mutated: got %d, want %d", v, int64(9007199254740993))
			}
		})
		t.Run("accepts exact power of two into int64", func(t *testing.T) {
			path, err := json.CreatePath("$.value")
			if err != nil {
				t.Fatal(err)
			}
			var v int64
			if err := path.Unmarshal([]byte(`{"value":1152921504606846976}`), &v); err != nil {
				t.Fatal(err)
			}
			if v != 1<<60 {
				t.Fatalf("decoded value = %d, want %d", v, int64(1<<60))
			}
		})
	})
}

func TestGetPath(t *testing.T) {
	t.Run("selector", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":{"b":10,"c":true},"b":"text"}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$.a.b")
		if err != nil {
			t.Fatal(err)
		}
		var b int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		if b != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("index", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":[{"b":10,"c":true},{"b":"text"}]}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$.a[0].b")
		if err != nil {
			t.Fatal(err)
		}
		var b int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		if b != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("indexAll", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":[{"b":1,"c":true},{"b":2},{"b":3}]}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$.a[*].b")
		if err != nil {
			t.Fatal(err)
		}
		var b []int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, []int{1, 2, 3}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("recursive", func(t *testing.T) {
		var v interface{}
		if err := json.Unmarshal([]byte(`{"a":[{"b":1,"c":true},{"b":2},{"b":3}],"a2":{"b":4}}`), &v); err != nil {
			t.Fatal(err)
		}
		path, err := json.CreatePath("$..b")
		if err != nil {
			t.Fatal(err)
		}
		var b []int
		if err := path.Get(v, &b); err != nil {
			t.Fatal(err)
		}
		sort.Ints(b)
		if !reflect.DeepEqual(b, []int{1, 2, 3, 4}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("rejects out of range integers", func(t *testing.T) {
		path, err := json.CreatePath("$.value")
		if err != nil {
			t.Fatal(err)
		}

		t.Run("signed width overflow", func(t *testing.T) {
			src := map[string]interface{}{"value": int64(2147483648)}
			v := int32(7)
			if err := path.Get(src, &v); err == nil {
				t.Fatal("expected error")
			}
			if v != 7 {
				t.Fatalf("destination mutated: got %d, want %d", v, 7)
			}
		})

		t.Run("unsigned to signed overflow", func(t *testing.T) {
			src := map[string]interface{}{"value": uint64(1 << 63)}
			v := int64(7)
			if err := path.Get(src, &v); err == nil {
				t.Fatal("expected error")
			}
			if v != 7 {
				t.Fatalf("destination mutated: got %d, want %d", v, 7)
			}
		})
		t.Run("accepts exact float64 integers", func(t *testing.T) {
			src := map[string]interface{}{"value": float64(1 << 53)}
			var v int64
			if err := path.Get(src, &v); err != nil {
				t.Fatal(err)
			}
			if v != 1<<53 {
				t.Fatalf("decoded value = %d, want %d", v, int64(1<<53))
			}
		})
		t.Run("accepts exact float64 power of two", func(t *testing.T) {
			src := map[string]interface{}{"value": float64(1 << 60)}
			var v int64
			if err := path.Get(src, &v); err != nil {
				t.Fatal(err)
			}
			if v != 1<<60 {
				t.Fatalf("decoded value = %d, want %d", v, int64(1<<60))
			}
		})
		t.Run("accepts exact float32 boundary", func(t *testing.T) {
			src := map[string]interface{}{"value": float32(1 << 24)}
			var v int64
			if err := path.Get(src, &v); err != nil {
				t.Fatal(err)
			}
			if v != 1<<24 {
				t.Fatalf("decoded value = %d, want %d", v, int64(1<<24))
			}
		})
		t.Run("accepts exact float64 just below uint64 sentinel", func(t *testing.T) {
			lower := math.Nextafter(math.Ldexp(1, 64), 0)
			src := map[string]interface{}{"value": lower}
			var v uint64
			if err := path.Get(src, &v); err != nil {
				t.Fatal(err)
			}
			if v != uint64(lower) {
				t.Fatalf("decoded value = %d, want %d", v, uint64(lower))
			}
		})
		t.Run("rejects uint64 sentinel float64", func(t *testing.T) {
			src := map[string]interface{}{"value": math.Ldexp(1, 64)}
			v := uint64(7)
			if err := path.Get(src, &v); err == nil {
				t.Fatal("expected error")
			}
			if v != 7 {
				t.Fatalf("destination mutated: got %d, want %d", v, uint64(7))
			}
		})
	})
	t.Run("supports struct traversal", func(t *testing.T) {
		type child struct {
			B int `json:"b"`
		}
		type root struct {
			A child `json:"a"`
		}

		path, err := json.CreatePath("$.a.b")
		if err != nil {
			t.Fatal(err)
		}

		var v int
		if err := path.Get(root{A: child{B: 10}}, &v); err != nil {
			t.Fatal(err)
		}
		if v != 10 {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("returns errors for nil pointer and interface branches", func(t *testing.T) {
		t.Run("nil pointer", func(t *testing.T) {
			type child struct {
				B int `json:"b"`
			}
			type root struct {
				A *child `json:"a"`
			}

			path, err := json.CreatePath("$.a.b")
			if err != nil {
				t.Fatal(err)
			}

			var v int
			if err := path.Get(root{}, &v); err == nil {
				t.Fatal("expected error")
			}
		})
		t.Run("nil interface", func(t *testing.T) {
			type root struct {
				A interface{} `json:"a"`
			}

			path, err := json.CreatePath("$.a.b")
			if err != nil {
				t.Fatal(err)
			}

			var v int
			if err := path.Get(root{}, &v); err == nil {
				t.Fatal("expected error")
			}
		})
	})
	t.Run("recursive traversal handles structs and nil branches", func(t *testing.T) {
		type child struct {
			B int `json:"b"`
		}
		type root struct {
			A *child  `json:"a"`
			C child   `json:"c"`
			D []child `json:"d"`
		}

		path, err := json.CreatePath("$..b")
		if err != nil {
			t.Fatal(err)
		}

		var v []int
		src := root{
			C: child{B: 2},
			D: []child{{B: 3}, {B: 1}},
		}
		if err := path.Get(src, &v); err != nil {
			t.Fatal(err)
		}
		sort.Ints(v)
		if !reflect.DeepEqual(v, []int{1, 2, 3}) {
			t.Fatalf("failed to decode by json.Get")
		}
	})
	t.Run("recursive assignment errors are returned", func(t *testing.T) {
		path, err := json.CreatePath("$..b")
		if err != nil {
			t.Fatal(err)
		}

		src := struct {
			A map[string]interface{} `json:"a"`
		}{
			A: map[string]interface{}{"b": 1.5},
		}

		var v int
		if err := path.Get(src, &v); err == nil {
			t.Fatal("expected error")
		}
		if v != 0 {
			t.Fatalf("destination mutated: got %d, want %d", v, 0)
		}
	})
	t.Run("rejects fractional floats for integers", func(t *testing.T) {
		path, err := json.CreatePath("$.value")
		if err != nil {
			t.Fatal(err)
		}

		var v int
		if err := path.Get(map[string]interface{}{"value": 1.5}, &v); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("rejects non-finite floats", func(t *testing.T) {
		path, err := json.CreatePath("$.value")
		if err != nil {
			t.Fatal(err)
		}

		for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			var v int64
			if err := path.Get(map[string]interface{}{"value": value}, &v); err == nil {
				t.Fatal("expected error")
			}
		}
	})
}
