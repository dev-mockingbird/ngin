package ngin

import "errors"

type Call func(ctx *Context, values ...Value) (bool, error)
type ValuedFunc func(ctx *Context) Value

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
	value   Value
}

func (v *Variable) Value() Value {
	if v.value != nil {
		return v.value
	}
	if v.Context == nil {
		return Null{}
	}
	if f, ok := v.Context.valuedFunks[v.Name]; ok {
		v.value = f(v.Context)
		return v.value
	}
	return v.Context.variables.Attr(v.Name)
}

func (v *Variable) Int() uint64 {
	return v.Value().Int()
}

func (v *Variable) Float() float64 {
	return v.Value().Float()
}

func (v *Variable) String() string {
	return v.Value().String()
}

func (v *Variable) Bytes() []byte {
	return v.Value().Bytes()
}

func (v *Variable) Bool() bool {
	return v.Value().Bool()
}

func (v *Variable) Slice() []Value {
	return []Value{v}
}

func (v Variable) Compare(r Value) int {
	return v.Value().Compare(r)
}

type Context struct {
	variables   Complex
	valuedFunks map[string]ValuedFunc
	funks       map[string]Call
	stmts       []Stmt
}

func NewContext() *Context {
	return &Context{
		variables:   Complex{attributes: make(map[string]*Complex)},
		valuedFunks: make(map[string]ValuedFunc),
		funks:       make(map[string]Call),
	}
}

func (ctx *Context) RegisterVariable(key string, val Value) {
	ctx.variables.SetAttr(key, val.Value())
}

func (ctx *Context) GetVariable(key string) Value {
	return ctx.variables.Attr(key)
}

func (ctx *Context) BindFunc(name string, funk Call) {
	ctx.funks[name] = funk
}

func (ctx *Context) BindValuedFunc(name string, funk ValuedFunc) {
	ctx.valuedFunks[name] = funk
}

func (ctx *Context) Call(name string, values ...Value) (bool, error) {
	if funk, ok := ctx.funks[name]; ok {
		return funk(ctx, values...)
	}
	return false, errors.New("call not found")
}

func (ctx *Context) NextStmts() []Stmt {
	return ctx.stmts
}
