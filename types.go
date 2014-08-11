package jigo

import (
	"fmt"
	"reflect"
)

// vartype is a simplified version of the notion of Kind in reflect, modified
// to reflect the slightly different semantics in jigo.
type vartype int

const (
	intType vartype = iota
	floatType
	stringType
	boolType
	sliceType
	mapType
	unknownType
)

func (v vartype) String() string {
	switch v {
	case intType:
		return "int"
	case floatType:
		return "float"
	case stringType:
		return "string"
	case boolType:
		return "bool"
	case sliceType:
		return "slice"
	case mapType:
		return "map"
	default:
		return "<unknown>"
	}
}

func isNumericVar(v vartype) bool {
	return v < stringType
}

func typeOf(i interface{}) vartype {
	switch i.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		return intType
	case float32, float64:
		return floatType
	case string:
		return stringType
	case bool:
		return boolType
	}
	kind := reflect.ValueOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		return sliceType
	case reflect.Map:
		return mapType
	}
	return unknownType
}

func asInteger(i interface{}) (int64, bool) {
	switch t := i.(type) {
	case uint:
		return int64(t), true
	case uint8:
		return int64(t), true
	case uint16:
		return int64(t), true
	case uint32:
		return int64(t), true
	case uint64:
		return int64(t), true
	case int:
		return int64(t), true
	case int8:
		return int64(t), true
	case int16:
		return int64(t), true
	case int32:
		return int64(t), true
	case int64:
		return t, true
	case float32:
		return int64(t), true
	case float64:
		return int64(t), true
	}
	return 0, false
}

func asFloat(i interface{}) (float64, bool) {
	switch t := i.(type) {
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case float32:
		return float64(t), true
	case float64:
		return t, true
	}
	return 0, false

}

func asString(i interface{}) string {
	return fmt.Sprint(i)
}
