package assert

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	intType   = reflect.TypeOf(int(1))
	int8Type  = reflect.TypeOf(int8(1))
	int16Type = reflect.TypeOf(int16(1))
	int32Type = reflect.TypeOf(int32(1))
	int64Type = reflect.TypeOf(int64(1))

	uintType   = reflect.TypeOf(uint(1))
	uint8Type  = reflect.TypeOf(uint8(1))
	uint16Type = reflect.TypeOf(uint16(1))
	uint32Type = reflect.TypeOf(uint32(1))
	uint64Type = reflect.TypeOf(uint64(1))

	float32Type = reflect.TypeOf(float32(1))
	float64Type = reflect.TypeOf(float64(1))

	stringType = reflect.TypeOf("")
)

const (
	compareLess         = compareFlag(0x01)
	compareEqual        = compareFlag(0x02)
	compareGreater      = compareFlag(0x04)
	compareLessEqual    = compareLess | compareEqual
	compareGreaterEqual = compareGreater | compareEqual
)

type compareFlag uint

func (f compareFlag) Is(target compareFlag) bool {
	return (f & target) != 0
}

// compare 比较两个object大小
func compare(obj1, obj2 interface{}, strict bool) (compareFlag, bool) {
	obj1Value := reflect.ValueOf(obj1)
	obj2Value := reflect.ValueOf(obj2)

	// throughout this switch we try and avoid calling .Convert() if possible,
	// as this has a pretty big performance impact
	switch obj1Value.Type().Kind() {
	case reflect.Int:
		{
			intobj1, ok := obj1.(int)
			if !ok {
				intobj1 = obj1Value.Convert(intType).Interface().(int)
			}
			intobj2, ok := obj2.(int)
			if !ok {
				intobj2 = obj2Value.Convert(intType).Interface().(int)
			}
			if intobj1 > intobj2 {
				return compareGreater, true
			}
			if intobj1 == intobj2 {
				return compareEqual, true
			}
			if intobj1 < intobj2 {
				return compareLess, true
			}
		}
	case reflect.Int8:
		{
			int8obj1, ok := obj1.(int8)
			if !ok {
				int8obj1 = obj1Value.Convert(int8Type).Interface().(int8)
			}
			int8obj2, ok := obj2.(int8)
			if !ok {
				int8obj2 = obj2Value.Convert(int8Type).Interface().(int8)
			}
			if int8obj1 > int8obj2 {
				return compareGreater, true
			}
			if int8obj1 == int8obj2 {
				return compareEqual, true
			}
			if int8obj1 < int8obj2 {
				return compareLess, true
			}
		}
	case reflect.Int16:
		{
			int16obj1, ok := obj1.(int16)
			if !ok {
				int16obj1 = obj1Value.Convert(int16Type).Interface().(int16)
			}
			int16obj2, ok := obj2.(int16)
			if !ok {
				int16obj2 = obj2Value.Convert(int16Type).Interface().(int16)
			}
			if int16obj1 > int16obj2 {
				return compareGreater, true
			}
			if int16obj1 == int16obj2 {
				return compareEqual, true
			}
			if int16obj1 < int16obj2 {
				return compareLess, true
			}
		}
	case reflect.Int32:
		{
			int32obj1, ok := obj1.(int32)
			if !ok {
				int32obj1 = obj1Value.Convert(int32Type).Interface().(int32)
			}
			int32obj2, ok := obj2.(int32)
			if !ok {
				int32obj2 = obj2Value.Convert(int32Type).Interface().(int32)
			}
			if int32obj1 > int32obj2 {
				return compareGreater, true
			}
			if int32obj1 == int32obj2 {
				return compareEqual, true
			}
			if int32obj1 < int32obj2 {
				return compareLess, true
			}
		}
	case reflect.Int64:
		{
			int64obj1, ok := obj1.(int64)
			if !ok {
				int64obj1 = obj1Value.Convert(int64Type).Interface().(int64)
			}
			int64obj2, ok := obj2.(int64)
			if !ok {
				int64obj2 = obj2Value.Convert(int64Type).Interface().(int64)
			}
			if int64obj1 > int64obj2 {
				return compareGreater, true
			}
			if int64obj1 == int64obj2 {
				return compareEqual, true
			}
			if int64obj1 < int64obj2 {
				return compareLess, true
			}
		}
	case reflect.Uint:
		{
			uintobj1, ok := obj1.(uint)
			if !ok {
				uintobj1 = obj1Value.Convert(uintType).Interface().(uint)
			}
			uintobj2, ok := obj2.(uint)
			if !ok {
				uintobj2 = obj2Value.Convert(uintType).Interface().(uint)
			}
			if uintobj1 > uintobj2 {
				return compareGreater, true
			}
			if uintobj1 == uintobj2 {
				return compareEqual, true
			}
			if uintobj1 < uintobj2 {
				return compareLess, true
			}
		}
	case reflect.Uint8:
		{
			uint8obj1, ok := obj1.(uint8)
			if !ok {
				uint8obj1 = obj1Value.Convert(uint8Type).Interface().(uint8)
			}
			uint8obj2, ok := obj2.(uint8)
			if !ok {
				uint8obj2 = obj2Value.Convert(uint8Type).Interface().(uint8)
			}
			if uint8obj1 > uint8obj2 {
				return compareGreater, true
			}
			if uint8obj1 == uint8obj2 {
				return compareEqual, true
			}
			if uint8obj1 < uint8obj2 {
				return compareLess, true
			}
		}
	case reflect.Uint16:
		{
			uint16obj1, ok := obj1.(uint16)
			if !ok {
				uint16obj1 = obj1Value.Convert(uint16Type).Interface().(uint16)
			}
			uint16obj2, ok := obj2.(uint16)
			if !ok {
				uint16obj2 = obj2Value.Convert(uint16Type).Interface().(uint16)
			}
			if uint16obj1 > uint16obj2 {
				return compareGreater, true
			}
			if uint16obj1 == uint16obj2 {
				return compareEqual, true
			}
			if uint16obj1 < uint16obj2 {
				return compareLess, true
			}
		}
	case reflect.Uint32:
		{
			uint32obj1, ok := obj1.(uint32)
			if !ok {
				uint32obj1 = obj1Value.Convert(uint32Type).Interface().(uint32)
			}
			uint32obj2, ok := obj2.(uint32)
			if !ok {
				uint32obj2 = obj2Value.Convert(uint32Type).Interface().(uint32)
			}
			if uint32obj1 > uint32obj2 {
				return compareGreater, true
			}
			if uint32obj1 == uint32obj2 {
				return compareEqual, true
			}
			if uint32obj1 < uint32obj2 {
				return compareLess, true
			}
		}
	case reflect.Uint64:
		{
			uint64obj1, ok := obj1.(uint64)
			if !ok {
				uint64obj1 = obj1Value.Convert(uint64Type).Interface().(uint64)
			}
			uint64obj2, ok := obj2.(uint64)
			if !ok {
				uint64obj2 = obj2Value.Convert(uint64Type).Interface().(uint64)
			}
			if uint64obj1 > uint64obj2 {
				return compareGreater, true
			}
			if uint64obj1 == uint64obj2 {
				return compareEqual, true
			}
			if uint64obj1 < uint64obj2 {
				return compareLess, true
			}
		}
	case reflect.Float32:
		{
			float32obj1, ok := obj1.(float32)
			if !ok {
				float32obj1 = obj1Value.Convert(float32Type).Interface().(float32)
			}
			float32obj2, ok := obj2.(float32)
			if !ok {
				float32obj2 = obj2Value.Convert(float32Type).Interface().(float32)
			}
			if float32obj1 > float32obj2 {
				return compareGreater, true
			}
			if float32obj1 == float32obj2 {
				return compareEqual, true
			}
			if float32obj1 < float32obj2 {
				return compareLess, true
			}
		}
	case reflect.Float64:
		{
			float64obj1, ok := obj1.(float64)
			if !ok {
				float64obj1 = obj1Value.Convert(float64Type).Interface().(float64)
			}
			float64obj2, ok := obj2.(float64)
			if !ok {
				float64obj2 = obj2Value.Convert(float64Type).Interface().(float64)
			}
			if float64obj1 > float64obj2 {
				return compareGreater, true
			}
			if float64obj1 == float64obj2 {
				return compareEqual, true
			}
			if float64obj1 < float64obj2 {
				return compareLess, true
			}
		}
	case reflect.String:
		{
			stringobj1, ok := obj1.(string)
			if !ok {
				stringobj1 = obj1Value.Convert(stringType).Interface().(string)
			}
			stringobj2, ok := obj2.(string)
			if !ok {
				stringobj2 = obj2Value.Convert(stringType).Interface().(string)
			}
			if stringobj1 > stringobj2 {
				return compareGreater, true
			}
			if stringobj1 == stringobj2 {
				return compareEqual, true
			}
			if stringobj1 < stringobj2 {
				return compareLess, true
			}
		}
	}

	return compareEqual, false
}

// isCompared 严格比较两个object
func isCompared(e1, e2 interface{}, expectedOrder compareFlag, strict bool) (bool, error) {
	e1Kind := reflect.ValueOf(e1).Kind()
	e2Kind := reflect.ValueOf(e2).Kind()
	if e1Kind != e2Kind {
		return false, fmt.Errorf("Elements should be the same type")
	}

	compareResult, isComparable := compare(e1, e2, strict)
	if !isComparable {
		return false, fmt.Errorf("Can not compare type \"%s\"", reflect.TypeOf(e1))
	}

	if !compareResult.Is(expectedOrder) {
		return false, nil
	}

	return true, nil
}

// isOrdered checks that collection contains orderable elements.
func isOrdered(object interface{}, expectedOrder compareFlag, strict bool) (bool, error) {
	objKind := reflect.TypeOf(object).Kind()
	if objKind != reflect.Slice && objKind != reflect.Array {
		return false, nil
	}

	objValue := reflect.ValueOf(object)
	objLen := objValue.Len()

	if objLen <= 1 {
		return true, nil
	}

	value := objValue.Index(0)
	valueInterface := value.Interface()

	for i := 1; i < objLen; i++ {
		prevValue := value
		prevValueInterface := valueInterface

		value = objValue.Index(i)
		valueInterface = value.Interface()

		compareResult, isComparable := compare(prevValueInterface, valueInterface, strict)

		if !isComparable {
			return false, fmt.Errorf("Can not compare type \"%s\" and \"%s\"", reflect.TypeOf(value), reflect.TypeOf(prevValue))
		}

		if !compareResult.Is(expectedOrder) {
			return false, nil
		}
	}

	return true, nil
}

// deepEqual 严格比较是否相等,类型+值
func deepEqual(obj1, obj2 interface{}) bool {
	if obj1 == nil || obj2 == nil {
		return obj1 == obj2
	}

	exp, ok := obj1.([]byte)
	if !ok {
		return reflect.DeepEqual(obj1, obj2)
	}

	act, ok := obj2.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}

	return bytes.Equal(exp, act)
}

// isDeepEqual 比较两个object是否相同,如果strict为true,则严格比较值和类型,否则仅比较值
func isDeepEqual(obj1, obj2 interface{}, strict bool) (bool, error) {
	if isFunc(obj1) || isFunc(obj2) {
		return false, fmt.Errorf("Invalid operation: %#v == %#v (cannot take func type as argument)", obj1, obj2)
	}

	if deepEqual(obj1, obj2) {
		return true, nil
	}

	if strict {
		return false, nil
	}

	type1 := reflect.TypeOf(obj1)
	if type1 == nil {
		return false, nil
	}
	value2 := reflect.ValueOf(obj2)
	if value2.IsValid() && value2.Type().ConvertibleTo(type1) {
		// Attempt comparison after type conversion
		return reflect.DeepEqual(value2.Convert(type1).Interface(), obj1), nil
	}

	return false, nil
}

// isFunc 判断是否是函数
func isFunc(arg interface{}) bool {
	if arg == nil {
		return false
	}

	return reflect.TypeOf(arg).Kind() == reflect.Func
}

// Stolen from the `go test` tool.
// isTest tells whether name looks like a test (or benchmark, according to prefix).
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer.
func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(r)
}

// samePointers compares two generic interface objects and returns whether
// they point to the same object
func isSamePointer(first, second interface{}) bool {
	firstPtr, secondPtr := reflect.ValueOf(first), reflect.ValueOf(second)
	if firstPtr.Kind() != reflect.Ptr || secondPtr.Kind() != reflect.Ptr {
		return false
	}

	firstType, secondType := reflect.TypeOf(first), reflect.TypeOf(second)
	if firstType != secondType {
		return false
	}

	// compare pointer addresses
	return first == second
}

// isNil checks if a specified object is nil or not, without Failing.
func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	isNilableKind := containsKind(
		[]reflect.Kind{
			reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map,
			reflect.Ptr, reflect.Slice},
		kind)

	if isNilableKind && value.IsNil() {
		return true
	}

	return false
}

// containsKind checks if a specified kind in the slice of kinds.
func containsKind(kinds []reflect.Kind, kind reflect.Kind) bool {
	for i := 0; i < len(kinds); i++ {
		if kind == kinds[i] {
			return true
		}
	}

	return false
}

// isEmpty gets whether the specified object is considered empty or not.
func isEmpty(object interface{}) bool {

	// get nil case out of the way
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
		// pointers are empty if nil or if the value they point to is empty
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

// isZero 判断是否是初始值
func isZero(object interface{}) bool {
	return object != nil && !reflect.DeepEqual(object, reflect.Zero(reflect.TypeOf(object)).Interface())
}

// getLen try to get length of object.
// return (false, 0) if impossible.
func getLen(x interface{}) (length int, err error) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%+v", e)
		}
	}()

	return v.Len(), nil
}

// containsElement try loop over the list check if the list includes the element.
// return (false, false) if impossible.
// return (true, false) if element was not found.
// return (true, true) if element was found.
func containsElement(list interface{}, element interface{}) (ok, found bool) {
	listValue := reflect.ValueOf(list)
	listKind := reflect.TypeOf(list).Kind()
	defer func() {
		if e := recover(); e != nil {
			ok = false
			found = false
		}
	}()

	if listKind == reflect.String {
		elementValue := reflect.ValueOf(element)
		return true, strings.Contains(listValue.String(), elementValue.String())
	}

	if listKind == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if deepEqual(mapKeys[i].Interface(), element) {
				return true, true
			}
		}
		return true, false
	}

	for i := 0; i < listValue.Len(); i++ {
		if deepEqual(listValue.Index(i).Interface(), element) {
			return true, true
		}
	}

	return true, false
}
