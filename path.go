package json

import (
	"reflect"

	"github.com/goccy/go-json/internal/decoder"
)

// CreatePath creates JSON Path.
//
// JSON Path rule
// $   : root object or element. The JSON Path format must start with this operator, which refers to the outermost level of the JSON-formatted string.
// .   : child operator. You can identify child values using dot-notation.
// ..  : recursive descent.
// []  : subscript operator. If the JSON object is an array, you can use brackets to specify the array index.
// [*] : all objects/elements for array.
//
// Reserved words must be properly escaped when included in Path.
//
// Escape Rule
// single quote style escape: e.g.) `$['a.b'].c`
// double quote style escape: e.g.) `$."a.b".c`
func CreatePath(p string) (*Path, error) {
	path, err := decoder.PathString(p).Build()
	if err != nil {
		return nil, err
	}
	return &Path{path: path}, nil
}

// Path represents JSON Path.
type Path struct {
	path *decoder.Path
}

// RootSelectorOnly whether only the root selector ($) is used.
func (p *Path) RootSelectorOnly() bool {
	return p.path.RootSelectorOnly
}

// UsedSingleQuotePathSelector whether single quote-based escaping was done when building the JSON Path.
func (p *Path) UsedSingleQuotePathSelector() bool {
	return p.path.SingleQuotePathSelector
}

// UsedSingleQuotePathSelector whether double quote-based escaping was done when building the JSON Path.
func (p *Path) UsedDoubleQuotePathSelector() bool {
	return p.path.DoubleQuotePathSelector
}

// Extract extracts a specific JSON string.
func (p *Path) Extract(data []byte, optFuncs ...DecodeOptionFunc) ([][]byte, error) {
	return extractFromPath(p, data, optFuncs...)
}

// PathString returns original JSON Path string.
func (p *Path) PathString() string {
	return p.path.String()
}

// Unmarshal extract and decode the value of the part corresponding to JSON Path from the input data.
func (p *Path) Unmarshal(data []byte, v interface{}, optFuncs ...DecodeOptionFunc) error {
	contents, err := extractFromPath(p, data, optFuncs...)
	if err != nil {
		return err
	}
	if handled, err := unmarshalPathExactIntegers(contents, reflect.ValueOf(v)); handled {
		return err
	}
	results := make([]interface{}, 0, len(contents))
	for _, content := range contents {
		var result interface{}
		if err := Unmarshal(content, &result); err != nil {
			return err
		}
		results = append(results, result)
	}
	if err := decoder.AssignValue(reflect.ValueOf(results), reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}

func unmarshalPathExactIntegers(contents [][]byte, dst reflect.Value) (bool, error) {
	if dst.Kind() != reflect.Ptr || dst.IsNil() {
		return false, nil
	}
	t := dst.Elem().Type()
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if len(contents) == 0 {
			return true, nil
		}
		ret := reflect.New(t)
		if err := Unmarshal(contents[0], ret.Interface()); err != nil {
			return true, err
		}
		dst.Elem().Set(ret.Elem())
		return true, nil
	case reflect.Slice:
		if !isExactIntegerKind(t.Elem().Kind()) {
			return false, nil
		}
		ret := reflect.MakeSlice(t, len(contents), len(contents))
		for i, content := range contents {
			if err := Unmarshal(content, ret.Index(i).Addr().Interface()); err != nil {
				return true, err
			}
		}
		dst.Elem().Set(ret)
		return true, nil
	case reflect.Array:
		if !isExactIntegerKind(t.Elem().Kind()) {
			return false, nil
		}
		if len(contents) != t.Len() {
			return true, decoder.AssignValue(reflect.ValueOf(make([]interface{}, len(contents))), dst)
		}
		ret := reflect.New(t).Elem()
		for i, content := range contents {
			if err := Unmarshal(content, ret.Index(i).Addr().Interface()); err != nil {
				return true, err
			}
		}
		dst.Elem().Set(ret)
		return true, nil
	default:
		return false, nil
	}
}

func isExactIntegerKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}

// Get extract and substitute the value of the part corresponding to JSON Path from the input value.
func (p *Path) Get(src, dst interface{}) error {
	return p.path.Get(reflect.ValueOf(src), reflect.ValueOf(dst))
}
