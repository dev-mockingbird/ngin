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
		a.SetAttr(attr[idx+1:], val)
		return
	}
	c.attributes[attr[:idx]] = &Complex{attributes: make(map[string]*Complex)}
	c.attributes[attr[:idx]].SetAttr(attr[idx+1:], val)
}

func (c *Complex) Attr(attr string) Value {
	return c.attr(attr, true)
}

func (c *Complex) AttrValue(attr string) Value {
	return c.attr(attr, false)
}

func (c *Complex) attr(attr string, key bool) Value {
	if len(attr) == 0 {
		return c.value
	}
	idx := strings.Index(attr, ".")
	if idx < 0 {
		if attr == "*" {
			ret := make(Slice, len(c.attributes))
			for k, v := range c.attributes {
				if key {
					ret = append(ret, String(k))
					continue
				}
				ret = append(ret, v.value)
			}
		}
		if a, ok := c.attributes[attr]; ok {
			return a.value
		}
		return Null{}
	}
	if a, ok := c.attributes[attr[:idx]]; ok {
		return a.attr(attr[idx+1:], key)
	}
	return Null{}
}
