package errorx

import (
	"errors"
	"fmt"
	"strings"
)

// Append is a helper function that will append more errors
// onto an Error in order to create a larger multi-error.
//
// If err is not a multierror.Error, then it will be turned into
// one. If any of the errs are multierr.Error, they will be flattened
// one level into err.
// Any nil errors within errs will be ignored. If err is nil, a new
// *Error will be returned.
func Append(err error, errs ...error) error {
	var merrs *MultiErr
	if err == nil {
		merrs = &MultiErr{}
	} else if m, ok := err.(*MultiErr); ok {
		merrs = m
	} else {
		merrs = &MultiErr{}
		merrs.Append(err)
	}

	merrs.Append(errs...)

	return merrs
}

// MultiErr 聚合多错误
type MultiErr struct {
	errors []error
}

func (e *MultiErr) Append(errs ...error) {
	e.errors = append(e.errors, errs...)
}

func (e *MultiErr) Errors() []error {
	return e.errors
}

func (e *MultiErr) Error() string {
	es := e.errors
	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", es[0])
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf("%d errors occurred:\n\t%s\n\n", len(es), strings.Join(points, "\n\t"))
}

func (e *MultiErr) Unwrap() error {
	if e == nil || len(e.errors) == 0 {
		return nil
	}

	// Shallow copy the slice
	errs := make([]error, len(e.errors))
	copy(errs, e.errors)
	return chain(errs)
}

// chain implements the interfaces necessary for errors.Is/As/Unwrap to
// work in a deterministic way with multierror. A chain tracks a list of
// errors while accounting for the current represented error. This lets
// Is/As be meaningful.
//
// Unwrap returns the next error. In the cleanest form, Unwrap would return
// the wrapped error here but we can't do that if we want to properly
// get access to all the errors. Instead, users are recommended to use
// Is/As to get the correct error type out.
//
// Precondition: []error is non-empty (len > 0)
type chain []error

// Error implements the error interface
func (e chain) Error() string {
	return e[0].Error()
}

// Unwrap implements errors.Unwrap by returning the next error in the
// chain or nil if there are no more errors.
func (e chain) Unwrap() error {
	if len(e) == 1 {
		return nil
	}

	return e[1:]
}

// As implements errors.As by attempting to map to the current value.
func (e chain) As(target interface{}) bool {
	return errors.As(e[0], target)
}

// Is implements errors.Is by comparing the current value directly.
func (e chain) Is(target error) bool {
	return errors.Is(e[0], target)
}
