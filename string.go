package ngin

import (
	"strconv"
	"strings"
)

type str struct {
	content string
}

func String(s string) Value {
	return str{content: s}
}

func (str str) WithContext(*Context) Value {
	return str
}

func (str str) Int() uint64 {
	ret, err := strconv.ParseUint(str.content, 10, 64)
	if err != nil {
		panic(err)
	}
	return ret
}

func (str str) Bool() bool {
	return strings.EqualFold(str.content, "true")
}

func (str str) Float() float64 {
	ret, err := strconv.ParseFloat(str.content, 64)
	if err != nil {
		panic(err)
	}
	return ret
}

func (str str) String() string {
	return str.content
}

func (str str) Bytes() []byte {
	return []byte(str.content)
}

func (str str) Slice() []Value {
	return []Value{str}
}

func (str str) Compare(val Value) int {
	return strings.Compare(str.content, val.String())
}

func (str str) Value() Value {
	return str
}
