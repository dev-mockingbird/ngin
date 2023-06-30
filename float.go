// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin

import "strconv"

type flt struct {
	value float64
}

func Float(value float64) Value {
	return flt{value: value}
}

func (f flt) Int() uint64 {
	return uint64(f.value)
}

func (f flt) Float() float64 {
	return f.value
}

func (f flt) String() string {
	return strconv.FormatFloat(f.value, 'f', 0, 64)
}

func (f flt) Bytes() []byte {
	return []byte(f.String())
}

func (f flt) Bool() bool {
	return f.Int() > 0
}

func (f flt) Compare(val Value) int {
	i := f.Int()
	j := val.Int()
	if i > j {
		return 1
	} else if i == j {
		return 0
	}
	return -1
}

func (f flt) Slice() []Value {
	return []Value{f}
}
