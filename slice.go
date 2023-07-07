package ngin

type Slice []Value

func (s Slice) Slice() []Value {
	return s
}

func (s Slice) WithContext(ctx *Context) Value {
	ret := make(Slice, len(s))
	for i, v := range s {
		ret[i] = v.WithContext(ctx)
	}
	return ret
}

func (s Slice) Int() uint64 {
	panic("can't get int value of slice")
}

func (s Slice) Float() float64 {
	panic("can't get float value of slice")
}

func (s Slice) String() string {
	panic("can't get string value of slice")
}

func (s Slice) Bytes() []byte {
	panic("can't get bytes value of slice")
}

func (s Slice) Bool() bool {
	panic("can't get bool value of slice")
}

func (s Slice) Contain(val Value) bool {
	for _, i := range s {
		if i.Compare(val) == 0 {
			return true
		}
	}
	return false
}

func (s Slice) Value() Value {
	return s
}

func (s Slice) Compare(val Value) int {
	r := val.Slice()
	if len(s) > len(r) {
		return 1
	} else if len(s) < len(r) {
		return -1
	}
	for i := 0; i < len(s); i++ {
		if ret := s[i].Compare(r[i]); ret != 0 {
			return ret
		}
	}
	return 0
}
