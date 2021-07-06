package assert

// T reports when failures occur.
// testing.T implements this interface.
type T interface {
	// Fail indicates that the test has failed but
	// allowed execution to continue.
	Fail()
	// FailNow indicates that the test has failed and
	// aborts the test.
	// FailNow is called in strict mode (via New).
	FailNow()
}

// New .
func New(t T, opts ...Option) *Assertion {
	a := &Assertion{t: t}
	if len(opts) > 0 {
		a.opts = newOptions(opts...)
	}
	return a
}

// Assertions provides assertion methods around the TestingT interface.
type Assertion struct {
	t    T
	opts *Options
}

func (a *Assertion) Positive(expected interface{}, argAndOpts ...interface{}) bool {
	return Positive(a.t, expected, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Negative(expected interface{}, argAndOpts ...interface{}) bool {
	return Negative(a.t, expected, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Greater(expected, actual interface{}, argAndOpts ...interface{}) bool {
	return Greater(a.t, expected, actual, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) GreaterEqual(expected, actual interface{}, argAndOpts ...interface{}) bool {
	return GreaterEqual(a.t, expected, actual, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Less(expected, actual interface{}, argAndOpts ...interface{}) bool {
	return Less(a.t, expected, actual, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) LessEqual(expected, actual interface{}, argAndOpts ...interface{}) bool {
	return LessEqual(a.t, expected, actual, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) IsIncreasing(expected interface{}, argAndOpts ...interface{}) bool {
	return IsIncreasing(a.t, expected, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) IsNonIncreasing(expected interface{}, argAndOpts ...interface{}) bool {
	return IsNonIncreasing(a.t, expected, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) IsDecreasing(expected interface{}, argAndOpts ...interface{}) bool {
	return IsDecreasing(a.t, expected, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) IsNonDecreasing(expected interface{}, argAndOpts ...interface{}) bool {
	return IsNonDecreasing(a.t, expected, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) IsType(expected, object interface{}, argAndOpts ...interface{}) bool {
	return IsType(a.t, expected, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Equal(expected, actual interface{}, argAndOpts ...interface{}) bool {
	return Equal(a.t, expected, actual, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NotEqual(expected, actual interface{}, argAndOpts ...interface{}) bool {
	return NotEqual(a.t, expected, actual, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Same(expected, object interface{}, argAndOpts ...interface{}) bool {
	return Same(a.t, expected, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NotSame(expected, object interface{}, argAndOpts ...interface{}) bool {
	return NotSame(a.t, expected, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Nil(object interface{}, argAndOpts ...interface{}) bool {
	return Nil(a.t, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NotNil(object interface{}, argAndOpts ...interface{}) bool {
	return NotNil(a.t, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Zero(object interface{}, argAndOpts ...interface{}) bool {
	return Nil(a.t, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NotZero(object interface{}, argAndOpts ...interface{}) bool {
	return NotNil(a.t, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Empty(object interface{}, argAndOpts ...interface{}) bool {
	return Empty(a.t, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NotEmpty(object interface{}, argAndOpts ...interface{}) bool {
	return NotEmpty(a.t, object, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Len(object interface{}, length int, argAndOpts ...interface{}) bool {
	return Len(a.t, object, length, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) True(value bool, argAndOpts ...interface{}) bool {
	return True(a.t, value, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) False(value bool, argAndOpts ...interface{}) bool {
	return False(a.t, value, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) Contains(s interface{}, contains interface{}, argAndOpts ...interface{}) bool {
	return Contains(a.t, s, contains, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) DirExists(path string, argAndOpts ...interface{}) bool {
	return DirExists(a.t, path, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NoDirExists(path string, argAndOpts ...interface{}) bool {
	return NoDirExists(a.t, path, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) FileExists(path string, argAndOpts ...interface{}) bool {
	return FileExists(a.t, path, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) NoFileExists(path string, argAndOpts ...interface{}) bool {
	return NoFileExists(a.t, path, a.wrapArgs(argAndOpts)...)
}

func (a *Assertion) wrapArgs(argAndOpts []interface{}) []interface{} {
	if a.opts == nil {
		return argAndOpts
	}

	return append(argAndOpts, withOptions(a.opts))
}
