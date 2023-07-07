package ngin

import (
	"bytes"
	"strconv"
)

type bs struct {
	content []byte
}

func Bytes(bytes []byte) Value {
	return bs{content: bytes}
}

func (bytes bs) WithContext(*Context) Value {
	return bytes
}

func (bytes bs) Int() uint64 {
	r, err := strconv.ParseUint(bytes.String(), 10, 64)
	if err != nil {
		panic(err)
	}
	return r
}

func (bytes bs) Float() float64 {
	return float64(bytes.Int())
}

func (bytes bs) String() string {
	return string(bytes.content)
}

func (bs bs) Bool() bool {
	return bytes.EqualFold(bs.content, []byte{'t', 'r', 'u', 'e'})
}

func (bytes bs) Bytes() []byte {
	return bytes.content
}

func (bs bs) Compare(val Value) int {
	return bytes.Compare(bs.content, val.Bytes())
}

func (bs bs) Slice() []Value {
	return []Value{bs}
}

func (bs bs) Value() Value {
	return bs
}
