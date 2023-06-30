// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin

type Bool struct {
	value bool
}

func (b Bool) Int() uint64 {
	if b.value {
		return 1
	}
	return 0
}

func (b Bool) Float() float64 {
	if b.value {
		return 1
	}
	return 0
}

func (b Bool) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b Bool) Bytes() []byte {
	if b.value {
		return []byte{'t', 'r', 'u', 'e'}
	}
	return []byte{'f', 'a', 'l', 's', 'e'}
}

func (b Bool) Slice() []Value {
	return []Value{b}
}

func (b Bool) Bool() bool {
	return b.value
}

func (b Bool) Compare(v Value) int {
	if b.value == v.Bool() {
		return 0
	} else if b.value {
		return 1
	} else {
		return -1
	}
}
