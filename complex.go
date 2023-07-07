package ngin

import (
	"encoding/json"
	"strings"
)

type Complex struct {
	attributes map[string]Value
}

func NewComplex() *Complex {
	return &Complex{attributes: make(map[string]Value)}
}

func (c *Complex) Value() Value {
	return c
}

func (c *Complex) Bool() bool {
	return false
}

func (c *Complex) Compare(val Value) int {
	panic("can't compare for complex")
}

func (c *Complex) Int() uint64 {
	panic("can't get int from complex")
}

func (c *Complex) Float() float64 {
	panic("can't get float from complex")
}

func (c *Complex) String() string {
	return string(c.Bytes())
}

func (c *Complex) Bytes() []byte {
	bs, err := json.Marshal(FromValue(c))
	if err != nil {
		panic(err)
	}
	return bs
}

func (c *Complex) WithContext(ctx *Context) Value {
	for _, v := range c.attributes {
		v.WithContext(ctx)
	}
	return c
}

func (c *Complex) Slice() []Value {
	return []Value{c}
}

func (c *Complex) SetAttr(attr string, val Value) {
	if len(attr) == 0 {
		return
	}
	idx := strings.Index(attr, ".")
	if idx < 0 {
		c.attributes[attr] = val
		return
	}
	a, ok := c.attributes[attr[:idx]]
	if !ok {
		c.attributes[attr[:idx]] = &Complex{attributes: make(map[string]Value)}
	}
	if _, ok := a.(*Complex); !ok {
		c.attributes[attr[:idx]] = &Complex{attributes: make(map[string]Value)}
	}
	c.attributes[attr[:idx]].(*Complex).SetAttr(attr[idx+1:], val)
}

func (c *Complex) Attr(attr string) Value {
	sub := c.find(attr)
	ret := []Value{}
	if s, ok := sub.(*Complex); ok {
		for k := range s.attributes {
			ret = append(ret, String(k))
		}
	}
	return Slice(ret)
}

func (c *Complex) AttrValue(attr string) Value {
	return c.find(attr)
}

func (c *Complex) find(attr string) Value {
	idx := strings.Index(attr, ".")
	current := attr
	last := ""
	if idx > -1 {
		current = attr[:idx]
		last = attr[idx+1:]
	}
	if current == "*" {
		ret := []Value{}
		for _, sub := range c.attributes {
			if last != "" {
				if s, ok := sub.(*Complex); ok {
					ret = append(ret, s.find(last))
				}
				continue
			}
			ret = append(ret, c)
		}
		return Slice(ret)
	}
	if sub, ok := c.attributes[current]; ok {
		if last != "" {
			if s, ok := sub.(*Complex); ok {
				return s.find(last)
			}
			return Null{}
		}
		return sub
	}
	return Null{}
}
