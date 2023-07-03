package ngin

import (
	"strings"

	"github.com/dev-mockingbird/logf"
)

type Func func(ctx *Context, values ...Value) (bool, error)
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
	if f := v.Context.GetValuedFunc(v.Name); f != nil {
		v.value = f(v.Context)
		return v.value
	}
	if v.Context.IsVar(v.Name) {
		return v.Context.GetValue(v.Name)
	}
	return String(v.Name)
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
	vars        map[string]struct{}
	funks       map[string]Func
	logger      logf.Logger
	bag         map[string]any
	stmts       []Stmt
	parent      *Context
}

func NewContext() *Context {
	return &Context{
		variables:   Complex{attributes: make(map[string]*Complex)},
		vars:        make(map[string]struct{}),
		valuedFunks: make(map[string]ValuedFunc),
		bag:         make(map[string]any),
		logger:      logf.New(),
		funks: map[string]Func{
			"var": defineVar,
		},
	}
}

func (ctx *Context) Folk() *Context {
	child := NewContext()
	child.parent = ctx
	child.stmts = ctx.stmts
	return child
}

func (ctx *Context) SetLogger(logger logf.Logger) {
	ctx.logger = logger
}

func (ctx *Context) Logger() logf.Logfer {
	return ctx.logger
}

func (ctx *Context) BindValue(key string, val Value) {
	ctx.variables.SetAttr(key, val.Value())
	if idx := strings.Index(key, "."); idx > -1 {
		if _, ok := ctx.vars[key[:idx]]; !ok {
			ctx.vars[key[:idx]] = struct{}{}
		}
		return
	}
	if _, ok := ctx.vars[key]; !ok {
		ctx.vars[key] = struct{}{}
	}
}

func (ctx *Context) Put(name string, val any) {
	ctx.bag[name] = val
}

func (ctx *Context) Get(name string) any {
	return ctx.bag[name]
}

func defineVar(ctx *Context, args ...Value) (bool, error) {
	for _, arg := range args {
		if v, ok := arg.(*Variable); ok {
			if ctx.IsVar(v.Name) {
				continue
			}
			ctx.vars[v.Name] = struct{}{}
		}
	}
	return true, nil
}

func (ctx *Context) IsVar(name string) bool {
	if ctx.isVar(name) {
		return true
	}
	if ctx.parent != nil {
		return ctx.parent.isVar(name)
	}
	return false
}

func (ctx *Context) isVar(name string) bool {
	if idx := strings.Index(name, "."); idx > -1 {
		_, ok := ctx.vars[name[:idx]]
		return ok
	}
	_, ok := ctx.vars[name]
	return ok
}

func (ctx *Context) GetValue(key string) Value {
	v := ctx.getValue(key)
	if _, ok := v.(Null); !ok && v != nil {
		return v
	}
	if ctx.parent != nil {
		return ctx.parent.GetValue(key)
	}
	return Null{}
}

func (ctx *Context) GetAttr(key string) Value {
	v := ctx.getAttr(key)
	if _, ok := v.(Null); ok {
		return v
	}
	if ctx.parent != nil {
		return ctx.parent.GetAttr(key)
	}
	return Null{}
}

func (ctx *Context) GetValuedFunc(name string) ValuedFunc {
	if funk, ok := ctx.valuedFunks[name]; ok {
		return funk
	}
	if ctx.parent != nil {
		return ctx.parent.GetValuedFunc(name)
	}
	return nil
}

func (ctx *Context) GetFunc(name string) Func {
	if funk, ok := ctx.funks[name]; ok {
		return funk
	}
	if ctx.parent != nil {
		return ctx.parent.GetFunc(name)
	}
	return nil
}

func (ctx *Context) getValue(key string) Value {
	return ctx.variables.AttrValue(key)
}

func (ctx *Context) getAttr(key string) Value {
	return ctx.variables.Attr(key)
}

func (ctx *Context) BindFunc(name string, funk Func) {
	ctx.funks[name] = funk
}

func (ctx *Context) BindValuedFunc(name string, funk ValuedFunc) {
	ctx.valuedFunks[name] = funk
}

func (ctx *Context) NextStmts() []Stmt {
	return ctx.stmts
}
