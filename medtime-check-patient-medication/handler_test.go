package function

import (
	"testing"
)

type TestHandlerI interface {
	GetAsserts() []Asserts
	GetBenchmarkRequest() Asserts
}

func NewAssert(f FunctionAssert) TestHandlerI {
	return f
}

func TestHandle(t *testing.T) {

}

func BenchmarkHandler(b *testing.B) {

}
