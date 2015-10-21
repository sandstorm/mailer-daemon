package router

import (
	"testing"
)

type Test struct {
	test *testing.T
}

func (this *Test) AssertEquals(message string, expected, actual interface{}) {
	if actual != expected {
		this.test.Errorf(message + ": expected '%v' is different from actual '%v'  ", expected, actual)
	}
}

func (this *Test) AssertTrue(message string, expected bool) {
	if ! expected {
		this.test.Error(message)
	}
}

func (this *Test) AssertNil(message string, actual interface{}) {
	if actual != nil {
		this.test.Errorf("%s: %+v is not nil", message, actual)
	}
}

func (this *Test) AssertFalse(message string, expected bool) {
	this.AssertTrue(message, ! expected)
}