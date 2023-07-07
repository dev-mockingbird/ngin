// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin

type bol struct {
	value bool
}

func Bool(value bool) Value {
	return bol{value: value}
}

func (b bol) WithContext(*Context) Value {
	return b
}

func (b bol) Int() uint64 {
	if b.value {
		return 1
	}
	return 0
}

func (b bol) Float() float64 {
	if b.value {
		return 1
	}
	return 0
}

func (b bol) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b bol) Bytes() []byte {
	if b.value {
		return []byte{'t', 'r', 'u', 'e'}
	}
	return []byte{'f', 'a', 'l', 's', 'e'}
}

func (b bol) Slice() []Value {
	return []Value{b}
}

func (b bol) Bool() bool {
	return b.value
}

func (b bol) Compare(v Value) int {
	if b.value == v.Bool() {
		return 0
	} else if b.value {
		return 1
	} else {
		return -1
	}
}

func (b bol) Value() Value {
	return b
}
