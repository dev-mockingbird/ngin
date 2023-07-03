package ngin_test

import (
	"testing"

	"github.com/dev-mockingbird/ngin"
)

func TestComplex_Set(t *testing.T) {
	complex := ngin.NewComplex()
	complex.SetAttr("", ngin.Int(1))
	value := complex.AttrValue("")
	if value.Int() != 1 {
		t.Fatal("root")
	}
	complex.SetAttr("hello.world", ngin.Int(1))
	value = complex.AttrValue("hello.world")
	if value.Int() != 1 {
		t.Fatal("hello.world")
	}
}
