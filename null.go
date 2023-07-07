// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin

type Null struct{}

func (Null) String() string {
	return ""
}

func (Null) Bytes() []byte {
	return nil
}

func (Null) Int() uint64 {
	return 0
}

func (n Null) WithContext(*Context) Value {
	return n
}

func (Null) Float() float64 {
	return 0
}

func (Null) Bool() bool {
	return false
}

func (n Null) Slice() []Value {
	return []Value{n}
}

func (n Null) Compare(v Value) int {
	if _, ok := v.(Null); ok {
		return 0
	}
	return -1
}

func (n Null) Value() Value {
	return n
}
