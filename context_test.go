// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin_test

import (
	"encoding/json"
	"testing"

	"github.com/dev-mockingbird/ngin"
)

func TestFromValue(t *testing.T) {
	c := ngin.NewComplex()
	c.SetAttr("hello.world", ngin.String("hello world"))
	c.SetAttr("hello.world1", ngin.String("hello world 1"))
	c.SetAttr("hello.world2", ngin.Bytes([]byte("hello world 2")))
	c.SetAttr("hello.world3", ngin.Int(123))
	c.SetAttr("hello.world4", ngin.Slice([]ngin.Value{ngin.String("h")}))

	result := ngin.FromValue(c)
	bs, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `{"hello":{"world":"hello world","world1":"hello world 1","world2":"hello world 2","world3":123,"world4":["h"]}}` {
		t.Fatal("not equal")
	}
}

func TestToValue(t *testing.T) {
	v := map[string]any{
		"hello": map[string]any{
			"world":  "hello world",
			"world1": "hello world 1",
			"world2": "hello world 2",
			"world3": 123,
			"world4": []string{"h"},
		},
	}
	result := ngin.ToValue(v)
	c, ok := result.(*ngin.Complex)
	if !ok {
		t.Fatal("type error")
	}
	if c.AttrValue("hello.world").String() != "hello world" ||
		c.AttrValue("hello.world1").String() != "hello world 1" ||
		c.AttrValue("hello.world2").String() != "hello world 2" {

		t.Fatal("string error")
	}
	if c.AttrValue("hello.world3").Int() != 123 {
		t.Fatal("int error")
	}
	_, ok = c.AttrValue("hello.world4").(ngin.Slice)
	if !ok {
		t.Fatal("slice error")
	}
}
