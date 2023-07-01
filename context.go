package ngin

import "errors"

type Call func(ctx *Context, values ...Value) (bool, error)

type Value interface {
	Int() uint64
	Float() float64
	String() string
	Bytes() []byte
	Bool() bool
	Slice() []Value
	Compare(Value) int
	Value() Value
}

type Variable struct {
	Name    string
	Context *Context
}

func (v Variable) Value() Value {
	return v.Context.variables.Attr(v.Name)
}

func (v Variable) Int() uint64 {
	return v.Value().Int()
}

func (v Variable) Float() float64 {
	return v.Value().Float()
}

func (v Variable) String() string {
	return v.Value().String()
}

func (v Variable) Bytes() []byte {
	return v.Value().Bytes()
}

func (v Variable) Bool() bool {
	return v.Value().Bool()
}

func (v Variable) Slice() []Value {
	return []Value{v}
}

func (v Variable) Compare(r Value) int {
	return v.Value().Compare(r)
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

func (ctx *Context) RegisterFunc(name string, funk Call) {
	ctx.funks[name] = funk
}

func (ctx *Context) Call(name string, values ...Value) (bool, error) {
	if funk, ok := ctx.funks[name]; ok {
		return funk(ctx, values...)
	}
	return false, errors.New("call not found")
}
