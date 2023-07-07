package ngin

import (
	"strconv"
)

type it struct {
	value uint64
}

func Int(v uint64) Value {
	return it{value: v}
}

func (it it) WithContext(*Context) Value {
	return it
}

func (it it) Int() uint64 {
	return it.value
}

func (it it) Bool() bool {
	return it.value > 0
}

func (it it) Float() float64 {
	return float64(it.value)
}

func (it it) Bytes() []byte {
	return []byte(it.String())
}

func (it it) String() string {
	return strconv.FormatUint(it.value, 10)
}

func (it it) Compare(val Value) int {
	r := val.Int()
	if it.value > r {
		return 1
	} else if it.value == r {
		return 0
	} else {
		return -1
	}
}

func (it it) Slice() []Value {
	return []Value{it}
}

func (it it) Value() Value {
	return it
}
