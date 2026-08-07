package decoder

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

var (
	nilValue = reflect.ValueOf(nil)
)

const (
	maxExactFloat32Integer = 1 << 24
	maxExactFloat64Integer = 1 << 53
)

func AssignValue(src, dst reflect.Value) error {
	if dst.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("invalid dst type. required pointer type: %T", dst.Type())
	}
	casted, err := castValue(dst.Elem().Type(), src)
	if err != nil {
		return err
	}
	dst.Elem().Set(casted)
	return nil
}

func castValue(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Int:
		return castSignedValue(t, v)
	case reflect.Int8:
		return castSignedValue(t, v)
	case reflect.Int16:
		return castSignedValue(t, v)
	case reflect.Int32:
		return castSignedValue(t, v)
	case reflect.Int64:
		return castSignedValue(t, v)
	case reflect.Uint:
		return castUnsignedValue(t, v)
	case reflect.Uint8:
		return castUnsignedValue(t, v)
	case reflect.Uint16:
		return castUnsignedValue(t, v)
	case reflect.Uint32:
		return castUnsignedValue(t, v)
	case reflect.Uint64:
		return castUnsignedValue(t, v)
	case reflect.Uintptr:
		return castUnsignedValue(t, v)
	case reflect.String:
		return castString(v)
	case reflect.Bool:
		return castBool(v)
	case reflect.Float32:
		vv, err := castFloat(v)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(float32(vv.Float())), nil
	case reflect.Float64:
		return castFloat(v)
	case reflect.Array:
		return castArray(t, v)
	case reflect.Slice:
		return castSlice(t, v)
	case reflect.Map:
		return castMap(t, v)
	case reflect.Struct:
		return castStruct(t, v)
	}
	return v, nil
}

func castSignedValue(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	vv, err := castInt(v)
	if err != nil {
		return nilValue, err
	}
	ret := reflect.New(t).Elem()
	i64 := vv.Int()
	if ret.OverflowInt(i64) {
		return nilValue, fmt.Errorf("value %d overflows %s", i64, t)
	}
	ret.SetInt(i64)
	return ret, nil
}

func castUnsignedValue(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	vv, err := castUint(v)
	if err != nil {
		return nilValue, err
	}
	ret := reflect.New(t).Elem()
	u64 := vv.Uint()
	if ret.OverflowUint(u64) {
		return nilValue, fmt.Errorf("value %d overflows %s", u64, t)
	}
	ret.SetUint(u64)
	return ret, nil
}

func castInt(v reflect.Value) (reflect.Value, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u64 := v.Uint()
		if u64 > math.MaxInt64 {
			return nilValue, fmt.Errorf("value %d overflows int64", u64)
		}
		return reflect.ValueOf(int64(u64)), nil
	case reflect.String:
		i64, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(i64), nil
	case reflect.Bool:
		if v.Bool() {
			return reflect.ValueOf(int64(1)), nil
		}
		return reflect.ValueOf(int64(0)), nil
	case reflect.Float32, reflect.Float64:
		i64, err := exactInt64FromFloat(v)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(i64), nil
	case reflect.Array:
		if v.Len() > 0 {
			return castInt(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to int64 from empty array")
	case reflect.Slice:
		if v.Len() > 0 {
			return castInt(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to int64 from empty slice")
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to int64 from nil interface")
		}
		return castInt(reflect.ValueOf(v.Interface()))
	case reflect.Map:
		return nilValue, fmt.Errorf("failed to cast to int64 from map")
	case reflect.Struct:
		return nilValue, fmt.Errorf("failed to cast to int64 from struct")
	case reflect.Ptr:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to int64 from nil pointer")
		}
		return castInt(v.Elem())
	}
	return nilValue, fmt.Errorf("failed to cast to int64 from %s", v.Type().Kind())
}

func castUint(v reflect.Value) (reflect.Value, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64 := v.Int()
		if i64 < 0 {
			return nilValue, fmt.Errorf("value %d overflows uint64", i64)
		}
		return reflect.ValueOf(uint64(i64)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v, nil
	case reflect.String:
		u64, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(u64), nil
	case reflect.Bool:
		if v.Bool() {
			return reflect.ValueOf(uint64(1)), nil
		}
		return reflect.ValueOf(uint64(0)), nil
	case reflect.Float32, reflect.Float64:
		u64, err := exactUint64FromFloat(v)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(u64), nil
	case reflect.Array:
		if v.Len() > 0 {
			return castUint(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to uint64 from empty array")
	case reflect.Slice:
		if v.Len() > 0 {
			return castUint(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to uint64 from empty slice")
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to uint64 from nil interface")
		}
		return castUint(reflect.ValueOf(v.Interface()))
	case reflect.Map:
		return nilValue, fmt.Errorf("failed to cast to uint64 from map")
	case reflect.Struct:
		return nilValue, fmt.Errorf("failed to cast to uint64 from struct")
	case reflect.Ptr:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to uint64 from nil pointer")
		}
		return castUint(v.Elem())
	}
	return nilValue, fmt.Errorf("failed to cast to uint64 from %s", v.Type().Kind())
}

func exactInt64FromFloat(v reflect.Value) (int64, error) {
	f64 := v.Float()
	if math.IsNaN(f64) || math.IsInf(f64, 0) || f64 < math.MinInt64 || f64 > math.MaxInt64 {
		return 0, fmt.Errorf("value %v overflows int64", f64)
	}
	if math.Trunc(f64) != f64 {
		return 0, fmt.Errorf("value %v is not an exact integer", f64)
	}
	if math.Abs(f64) >= maxExactIntegerFromFloat(v.Kind()) {
		return 0, fmt.Errorf("value %v is not an exact integer", f64)
	}
	return int64(f64), nil
}

func exactUint64FromFloat(v reflect.Value) (uint64, error) {
	f64 := v.Float()
	if math.IsNaN(f64) || math.IsInf(f64, 0) || f64 < 0 || f64 > float64(math.MaxUint64) {
		return 0, fmt.Errorf("value %v overflows uint64", f64)
	}
	if math.Trunc(f64) != f64 {
		return 0, fmt.Errorf("value %v is not an exact integer", f64)
	}
	if f64 >= maxExactIntegerFromFloat(v.Kind()) {
		return 0, fmt.Errorf("value %v is not an exact integer", f64)
	}
	return uint64(f64), nil
}

func maxExactIntegerFromFloat(kind reflect.Kind) float64 {
	if kind == reflect.Float32 {
		return maxExactFloat32Integer
	}
	return maxExactFloat64Integer
}

func castString(v reflect.Value) (reflect.Value, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(fmt.Sprint(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return reflect.ValueOf(fmt.Sprint(v.Uint())), nil
	case reflect.String:
		return v, nil
	case reflect.Bool:
		if v.Bool() {
			return reflect.ValueOf("true"), nil
		}
		return reflect.ValueOf("false"), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(fmt.Sprint(v.Float())), nil
	case reflect.Array:
		if v.Len() > 0 {
			return castString(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to string from empty array")
	case reflect.Slice:
		if v.Len() > 0 {
			return castString(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to string from empty slice")
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to string from nil interface")
		}
		return castString(reflect.ValueOf(v.Interface()))
	case reflect.Map:
		return nilValue, fmt.Errorf("failed to cast to string from map")
	case reflect.Struct:
		return nilValue, fmt.Errorf("failed to cast to string from struct")
	case reflect.Ptr:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to string from nil pointer")
		}
		return castString(v.Elem())
	}
	return nilValue, fmt.Errorf("failed to cast to string from %s", v.Type().Kind())
}

func castBool(v reflect.Value) (reflect.Value, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v.Int() {
		case 0:
			return reflect.ValueOf(false), nil
		case 1:
			return reflect.ValueOf(true), nil
		}
		return nilValue, fmt.Errorf("failed to cast to bool from %d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch v.Uint() {
		case 0:
			return reflect.ValueOf(false), nil
		case 1:
			return reflect.ValueOf(true), nil
		}
		return nilValue, fmt.Errorf("failed to cast to bool from %d", v.Uint())
	case reflect.String:
		b, err := strconv.ParseBool(v.String())
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(b), nil
	case reflect.Bool:
		return v, nil
	case reflect.Float32, reflect.Float64:
		switch v.Float() {
		case 0:
			return reflect.ValueOf(false), nil
		case 1:
			return reflect.ValueOf(true), nil
		}
		return nilValue, fmt.Errorf("failed to cast to bool from %f", v.Float())
	case reflect.Array:
		if v.Len() > 0 {
			return castBool(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to string from empty array")
	case reflect.Slice:
		if v.Len() > 0 {
			return castBool(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to string from empty slice")
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to bool from nil interface")
		}
		return castBool(reflect.ValueOf(v.Interface()))
	case reflect.Map:
		return nilValue, fmt.Errorf("failed to cast to string from map")
	case reflect.Struct:
		return nilValue, fmt.Errorf("failed to cast to string from struct")
	case reflect.Ptr:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to bool from nil pointer")
		}
		return castBool(v.Elem())
	}
	return nilValue, fmt.Errorf("failed to cast to bool from %s", v.Type().Kind())
}

func castFloat(v reflect.Value) (reflect.Value, error) {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(float64(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return reflect.ValueOf(float64(v.Uint())), nil
	case reflect.String:
		f64, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(f64), nil
	case reflect.Bool:
		if v.Bool() {
			return reflect.ValueOf(float64(1)), nil
		}
		return reflect.ValueOf(float64(0)), nil
	case reflect.Float32, reflect.Float64:
		return v, nil
	case reflect.Array:
		if v.Len() > 0 {
			return castFloat(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to float64 from empty array")
	case reflect.Slice:
		if v.Len() > 0 {
			return castFloat(v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to float64 from empty slice")
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to float64 from nil interface")
		}
		return castFloat(reflect.ValueOf(v.Interface()))
	case reflect.Map:
		return nilValue, fmt.Errorf("failed to cast to float64 from map")
	case reflect.Struct:
		return nilValue, fmt.Errorf("failed to cast to float64 from struct")
	case reflect.Ptr:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to float64 from nil pointer")
		}
		return castFloat(v.Elem())
	}
	return nilValue, fmt.Errorf("failed to cast to float64 from %s", v.Type().Kind())
}

func castArray(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	kind := v.Type().Kind()
	if kind == reflect.Interface {
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to array from nil interface")
		}
		return castArray(t, reflect.ValueOf(v.Interface()))
	}
	if kind != reflect.Slice && kind != reflect.Array {
		return nilValue, fmt.Errorf("failed to cast to array from %s", kind)
	}
	if t.Elem() == v.Type().Elem() {
		return v, nil
	}
	if t.Len() != v.Len() {
		return nilValue, fmt.Errorf("failed to cast [%d]array from slice of %d length", t.Len(), v.Len())
	}
	ret := reflect.New(t).Elem()
	for i := 0; i < v.Len(); i++ {
		vv, err := castValue(t.Elem(), v.Index(i))
		if err != nil {
			return nilValue, err
		}
		ret.Index(i).Set(vv)
	}
	return ret, nil
}

func castSlice(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	kind := v.Type().Kind()
	if kind == reflect.Interface {
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to slice from nil interface")
		}
		return castSlice(t, reflect.ValueOf(v.Interface()))
	}
	if kind != reflect.Slice && kind != reflect.Array {
		return nilValue, fmt.Errorf("failed to cast to slice from %s", kind)
	}
	if t.Elem() == v.Type().Elem() {
		return v, nil
	}
	ret := reflect.MakeSlice(t, v.Len(), v.Len())
	for i := 0; i < v.Len(); i++ {
		vv, err := castValue(t.Elem(), v.Index(i))
		if err != nil {
			return nilValue, err
		}
		ret.Index(i).Set(vv)
	}
	return ret, nil
}

func castMap(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	ret := reflect.MakeMap(t)
	switch v.Type().Kind() {
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key, err := castValue(t.Key(), iter.Key())
			if err != nil {
				return nilValue, err
			}
			value, err := castValue(t.Elem(), iter.Value())
			if err != nil {
				return nilValue, err
			}
			ret.SetMapIndex(key, value)
		}
		return ret, nil
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to map from nil interface")
		}
		return castMap(t, reflect.ValueOf(v.Interface()))
	case reflect.Slice:
		if v.Len() > 0 {
			return castMap(t, v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to map from empty slice")
	}
	return nilValue, fmt.Errorf("failed to cast to map from %s", v.Type().Kind())
}

func castStruct(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	ret := reflect.New(t).Elem()
	switch v.Type().Kind() {
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			k, err := castString(key)
			if err != nil {
				return nilValue, err
			}
			fieldName := k.String()
			field, ok := t.FieldByName(fieldName)
			if ok {
				value, err := castValue(field.Type, iter.Value())
				if err != nil {
					return nilValue, err
				}
				ret.FieldByName(fieldName).Set(value)
			}
		}
		return ret, nil
	case reflect.Struct:
		for i := 0; i < v.Type().NumField(); i++ {
			name := v.Type().Field(i).Name
			ret.FieldByName(name).Set(v.FieldByName(name))
		}
		return ret, nil
	case reflect.Interface:
		if v.IsNil() {
			return nilValue, fmt.Errorf("failed to cast to struct from nil interface")
		}
		return castStruct(t, reflect.ValueOf(v.Interface()))
	case reflect.Slice:
		if v.Len() > 0 {
			return castStruct(t, v.Index(0))
		}
		return nilValue, fmt.Errorf("failed to cast to struct from empty slice")
	default:
		return nilValue, fmt.Errorf("failed to cast to struct from %s", v.Type().Kind())
	}
}
