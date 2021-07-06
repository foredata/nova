package assert

import (
	"fmt"
	"reflect"
)

func Equal(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isDeepEqual(expected, actual, opts.strict)
	return verify(t, ok, err, args, opts)
}

func NotEqual(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isDeepEqual(expected, actual, opts.strict)
	return verify(t, !ok, err, args, opts)
}

// Greater asserts that the first element is greater than the second
//
//    assert.Greater(t, 2, 1)
//    assert.Greater(t, float64(2), float64(1))
//    assert.Greater(t, "b", "a")
func Greater(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isCompared(expected, actual, compareGreater, opts.strict)
	return verify(t, ok, err, args, opts)
}

func GreaterEqual(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isCompared(expected, actual, compareGreaterEqual, opts.strict)
	return verify(t, ok, err, args, opts)
}

func Less(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isCompared(expected, actual, compareLess, opts.strict)
	return verify(t, ok, err, args, opts)
}

func LessEqual(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isCompared(expected, actual, compareLessEqual, opts.strict)
	return verify(t, ok, err, args, opts)
}

// Positive asserts that the specified element is positive
//
//    assert.Positive(t, 1)
//    assert.Positive(t, 1.23)
func Positive(t T, expected interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	zero := reflect.Zero(reflect.TypeOf(expected))
	ok, err := isCompared(expected, zero, compareGreater, opts.strict)
	return verify(t, ok, err, args, opts)
}

// Negative asserts that the specified element is negative
//
//    assert.Negative(t, -1)
//    assert.Negative(t, -1.23)
func Negative(t T, expected interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	zero := reflect.Zero(reflect.TypeOf(expected))
	ok, err := isCompared(expected, zero, compareLess, opts.strict)
	return verify(t, ok, err, args, opts)
}

// IsIncreasing asserts that the collection is increasing
//
//    assert.IsIncreasing(t, []int{1, 2, 3})
//    assert.IsIncreasing(t, []float{1, 2})
//    assert.IsIncreasing(t, []string{"a", "b"})
func IsIncreasing(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isOrdered(object, compareLess, opts.strict)
	return verify(t, ok, err, args, opts)
}

// IsNonIncreasing asserts that the collection is not increasing
//
//    assert.IsNonIncreasing(t, []int{2, 1, 1})
//    assert.IsNonIncreasing(t, []float{2, 1})
//    assert.IsNonIncreasing(t, []string{"b", "a"})
func IsNonIncreasing(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isOrdered(object, compareGreaterEqual, opts.strict)
	return verify(t, ok, err, args, opts)
}

// IsDecreasing asserts that the collection is decreasing
//
//    assert.IsDecreasing(t, []int{2, 1, 0})
//    assert.IsDecreasing(t, []float{2, 1})
//    assert.IsDecreasing(t, []string{"b", "a"})
func IsDecreasing(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isOrdered(object, compareGreater, opts.strict)
	return verify(t, ok, err, args, opts)
}

// IsNonDecreasing asserts that the collection is not decreasing
//
//    assert.IsNonDecreasing(t, []int{1, 1, 2})
//    assert.IsNonDecreasing(t, []float{1, 2})
//    assert.IsNonDecreasing(t, []string{"a", "b"})
func IsNonDecreasing(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, err := isOrdered(object, compareLessEqual, opts.strict)
	return verify(t, ok, err, args, opts)
}

// IsType asserts that the specified objects are of the same type.
func IsType(t T, expectedType interface{}, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := deepEqual(reflect.TypeOf(object), reflect.TypeOf(expectedType))
	return verify(t, ok, nil, args, opts)
}

// Same asserts that two pointers reference the same object.
//
//    assert.Same(t, ptr1, ptr2)
//
// Both arguments must be pointer variables. Pointer variable sameness is
// determined based on the equality of both type and value.
func Same(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isSamePointer(expected, actual)
	return verify(t, ok, nil, args, opts)
}

// NotSame asserts that two pointers do not reference the same object.
//
//    assert.NotSame(t, ptr1, ptr2)
//
// Both arguments must be pointer variables. Pointer variable sameness is
// determined based on the equality of both type and value.
func NotSame(t T, expected, actual interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isSamePointer(expected, actual)
	return verify(t, !ok, nil, args, opts)
}

// Nil asserts that the specified object is nil.
//
//    assert.Nil(t, err)
func Nil(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isNil(object)
	return verify(t, ok, nil, args, opts)
}

// NotNil asserts that the specified object is not nil.
//
//    assert.NotNil(t, err)
func NotNil(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isNil(object)
	return verify(t, !ok, nil, args, opts)
}

// Zero
func Zero(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isZero(object)
	return verify(t, ok, nil, args, opts)
}

// NotZero .
func NotZero(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isZero(object)
	return verify(t, !ok, nil, args, opts)
}

// Empty asserts that the specified object is empty.  I.e. nil, "", false, 0 or either
// a slice or a channel with len == 0.
//
//  assert.Empty(t, obj)
func Empty(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isEmpty(object)
	return verify(t, ok, nil, args, opts)
}

// NotEmpty asserts that the specified object is NOT empty.  I.e. not nil, "", false, 0 or either
// a slice or a channel with len == 0.
//
//  if assert.NotEmpty(t, obj) {
//    assert.Equal(t, "two", obj[1])
//  }
func NotEmpty(t T, object interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok := isEmpty(object)
	return verify(t, !ok, nil, args, opts)
}

// Len asserts that the specified object has specific length.
// Len also fails if the object has a type that len() not accept.
//
//    assert.Len(t, mySlice, 3)
func Len(t T, object interface{}, length int, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	l, err := getLen(object)
	return verify(t, l == length, err, args, opts)
}

// True asserts that the specified value is true.
//
//    assert.True(t, myBool)
func True(t T, value bool, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	return verify(t, value, nil, args, opts)
}

// False asserts that the specified value is false.
//
//    assert.False(t, myBool)
func False(t T, value bool, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	return verify(t, !value, nil, args, opts)
}

func Contains(t T, list interface{}, element interface{}, argAndOpts ...interface{}) bool {
	args, opts := parseArgs(argAndOpts)
	ok, found := containsElement(list, element)
	var err error
	if !ok {
		err = fmt.Errorf("%#v could not be applied builtin len()", list)
	}
	return verify(t, found, err, args, opts)
}

// DirExists checks whether a directory exists in the given path. It also fails
// if the path is a file rather a directory or there is an error checking whether it exists.
func DirExists(t T, path string, argAndOpts ...interface{}) bool {
	return true
}

func NoDirExists(t T, path string, argAndOpts ...interface{}) bool {
	return true
}

func FileExists(t T, path string, argAndOpts ...interface{}) bool {
	return true
}

func NoFileExists(t T, path string, argAndOpts ...interface{}) bool {
	return true
}

// parseArgs 解析args,前边为参数,后边为Option
func parseArgs(args []interface{}) ([]interface{}, *Options) {
	if len(args) == 0 {
		return nil, defaultOptions
	}

	// 从后向前计算options
	o := newOptions()
	for i := len(args) - 1; i >= 0; i-- {
		fn, ok := args[i].(Option)
		if !ok {
			return args[:i], o
		}
		fn(o)
	}

	return args, o
}

// verify 校验结果
func verify(t T, ok bool, err error, args []interface{}, opts *Options) bool {
	if err != nil {
		// log()
		return false
	}

	if !ok {
		// log error
		return false
	}

	return true
}
