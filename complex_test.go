package ngin_test

import (
	"fmt"
	"testing"

	"github.com/dev-mockingbird/ngin"
)

func TestComplex_Set(t *testing.T) {
	complex := ngin.NewComplex()
	complex.SetAttr("root", ngin.Int(1))
	value := complex.AttrValue("root")
	if value.Int() != 1 {
		t.Fatal("root")
	}
	complex.SetAttr("hello.world", ngin.Int(1))
	value = complex.AttrValue("hello.world")
	if value.Int() != 1 {
		t.Fatal("hello.world")
	}
}

func TestComplex_Find(t *testing.T) {
	complex := ngin.NewComplex()
	for i := 0; i < 10; i++ {
		complex.SetAttr(fmt.Sprintf("hello.%d.world", i), ngin.Int(uint64(i)))
	}
	attrs := ngin.Slice(complex.Attr("hello").Slice())
	if len(attrs) != 10 {
		t.Fatal("attr length")
	}
	for i := 0; i < 10; i++ {
		if !attrs.Contain(ngin.Int(uint64(i))) {
			t.Fatalf("attr at %d", i)
		}
	}
	values := ngin.Slice(complex.AttrValue("hello.*.world").Slice())
	for i := 0; i < 10; i++ {
		if !values.Contain(ngin.Int(uint64(i))) {
			t.Fatalf("attr value at %d", i)
		}
	}
}
