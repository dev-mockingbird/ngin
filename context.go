package ngin

import "errors"

type Call func(values ...Value) (bool, error)

type Value interface {
	Int() uint64
	Float() float64
	String() string
	Bytes() []byte
	Bool() bool
	Slice() []Value
	Compare(Value) int
}

type Context struct {
	variables Complex
	funks     map[string]Call
}

func NewContext() *Context {
	return &Context{
		variables: Complex{attributes: make(map[string]*Complex)},
		funks:     make(map[string]Call),
	}
}

func (ctx *Context) RegisterVariable(key string, val Value) {
	ctx.variables.SetAttr(key, val)
}

func (ctx *Context) GetVariable(key string) Value {
	return ctx.variables.value
}

func (ctx *Context) RegisterFunc(name string, funk func(values ...Value) (bool, error)) {
	ctx.funks[name] = funk
}

func (ctx *Context) Call(name string, values ...Value) (bool, error) {
	if funk, ok := ctx.funks[name]; ok {
		return funk(values...)
	}
	return false, errors.New("call not found")
}
