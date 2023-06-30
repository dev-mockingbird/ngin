// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/dev-mockingbir/ngin"
)

func TestInt(t *testing.T) {
	i := ngin.Int(1234567890)
	if i.Int() != 1234567890 {
		t.Fatal("int")
	}
	s := i.String()
	if s != "1234567890" {
		t.Fatal("string")
	}
	if !i.Bool() {
		t.Fatal("bool")
	}
	if !bytes.Equal(i.Bytes(), []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}) {
		t.Fatal("bytes")
	}
	f := i.Float()
	if math.Round(f) != 1234567890 {
		t.Fatal("float")
	}
}
