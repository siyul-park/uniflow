package testing

import (
	"fmt"
	"reflect"
)

// Assertion provides methods for test assertions.
type Assertion struct {
	t *Tester
}

// NewAssertion creates a new Assertion instance.
func NewAssertion(t *Tester) *Assertion {
	return &Assertion{t: t}
}

// Equal asserts that two values are equal.
func (a *Assertion) Equal(expected, actual interface{}) bool {
	if !reflect.DeepEqual(expected, actual) {
		err := fmt.Errorf("expected %v, but got %v", expected, actual)
		a.t.Close(err)
		return false
	}
	return true
}

// NotEqual asserts that two values are not equal.
func (a *Assertion) NotEqual(expected, actual interface{}) bool {
	if reflect.DeepEqual(expected, actual) {
		err := fmt.Errorf("expected %v to be different from %v", expected, actual)
		a.t.Close(err)
		return false
	}
	return true
}

// True asserts that a value is true.
func (a *Assertion) True(value bool) bool {
	if !value {
		err := fmt.Errorf("expected true, but got false")
		a.t.Close(err)
		return false
	}
	return true
}

// False asserts that a value is false.
func (a *Assertion) False(value bool) bool {
	if value {
		err := fmt.Errorf("expected false, but got true")
		a.t.Close(err)
		return false
	}
	return true
}

// Nil asserts that a value is nil.
func (a *Assertion) Nil(value interface{}) bool {
	if value != nil && !reflect.ValueOf(value).IsNil() {
		err := fmt.Errorf("expected nil, but got %v", value)
		a.t.Close(err)
		return false
	}
	return true
}

// NotNil asserts that a value is not nil.
func (a *Assertion) NotNil(value interface{}) bool {
	if value == nil || reflect.ValueOf(value).IsNil() {
		err := fmt.Errorf("expected non-nil value")
		a.t.Close(err)
		return false
	}
	return true
}

// NoError asserts that an error is nil.
func (a *Assertion) NoError(err error) bool {
	if err != nil {
		a.t.Close(fmt.Errorf("expected no error, but got: %v", err))
		return false
	}
	return true
}

// Error asserts that an error is not nil.
func (a *Assertion) Error(err error) bool {
	if err == nil {
		a.t.Close(fmt.Errorf("expected an error"))
		return false
	}
	return true
}
