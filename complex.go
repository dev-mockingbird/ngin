package ngin

import "strings"

type Complex struct {
	attributes map[string]*Complex
	value      Value
}

func NewComplex() *Complex {
	return &Complex{attributes: make(map[string]*Complex), value: Null{}}
}

func (c *Complex) Value() Value {
	return c.value
}

func (c *Complex) SetAttr(attr string, val Value) {
	if len(attr) == 0 {
		c.value = val
		return
	}
	idx := strings.Index(attr, ".")
	if idx < 0 {
		if a, ok := c.attributes[attr]; ok {
			a.value = val
			return
		}
		c.attributes[attr] = &Complex{value: val, attributes: make(map[string]*Complex)}
		return
	}
	if a, ok := c.attributes[attr[:idx]]; ok {
		a.SetAttr(attr[idx:], val)
		return
	}
	c.attributes[attr[:idx]] = &Complex{attributes: make(map[string]*Complex)}
	c.attributes[attr[:idx]].SetAttr(attr[idx+1:], val)
}

func (c *Complex) Attr(attr string) Value {
	if len(attr) == 0 {
		return c.value
	}
	idx := strings.Index(attr, ".")
	if idx < 0 {
		if a, ok := c.attributes[attr]; ok {
			return a.value
		}
		return Null{}
	}
	if a, ok := c.attributes[attr[:idx]]; ok {
		return a.Attr(attr[idx+1:])
	}
	return Null{}
}
