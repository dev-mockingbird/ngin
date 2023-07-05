package ngin

import (
	"reflect"
	"strings"

	"github.com/dev-mockingbird/logf"
)

type Func func(ctx *Context, values ...Value) (bool, error)
type ValuedFunc func(ctx *Context, values ...Value) Value

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

func ToValue(m any) Value {
	rv := reflect.ValueOf(m)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		ret := []Value{}
		for i := 0; i < rv.Len(); i++ {
			val := rv.Index(i)
			if val.CanInterface() {
				ret = append(ret, ToValue(val.Interface()))
			}
		}
		return Slice(ret)
	case reflect.Bool:
		return Bool(rv.Interface().(bool))
	case reflect.Float32:
		return Float(float64(rv.Interface().(float32)))
	case reflect.Float64:
		return Float(rv.Interface().(float64))
	case reflect.Uint:
		return Int(uint64(uint64(rv.Interface().(uint))))
	case reflect.Uint8:
		return Int(uint64(uint64(rv.Interface().(uint8))))
	case reflect.Uint16:
		return Int(uint64(rv.Interface().(uint16)))
	case reflect.Uint32:
		return Int(uint64(rv.Interface().(uint32)))
	case reflect.Uint64:
		return Int(uint64(rv.Interface().(uint64)))
	case reflect.Int:
		return Int(uint64(uint64(rv.Interface().(int))))
	case reflect.Int8:
		return Int(uint64(uint64(rv.Interface().(int8))))
	case reflect.Int16:
		return Int(uint64(rv.Interface().(int16)))
	case reflect.Int32:
		return Int(uint64(rv.Interface().(int32)))
	case reflect.Int64:
		return Int(uint64(rv.Interface().(int64)))
	case reflect.String:
		return String(rv.Interface().(string))
	case reflect.Map:
		ret := NewComplex()
		for _, k := range rv.MapKeys() {
			if k.Kind() != reflect.String {
				continue
			}
			val := rv.MapIndex(k)
			if !val.CanInterface() || !k.CanInterface() {
				continue
			}
			ret.SetAttr(k.Interface().(string), ToValue(val.Interface()))
		}
		return ret
	case reflect.Struct:
		ret := NewComplex()
		for i := 0; i < rv.NumField(); i++ {
			k := rv.Type().Field(i).Name
			fv := rv.Field(i)
			if !fv.CanInterface() {
				continue
			}
			ret.SetAttr(k, ToValue(fv.Interface()))
		}
		return ret
	}
	return nil
}

func FromValue(v Value) any {
	if c, ok := v.(*Complex); ok {
		ret := make(map[string]any)
		for k, val := range c.attributes {
			ret[k] = FromValue(val)
		}
		return ret
	} else if s, ok := v.(Slice); ok {
		ret := make([]any, len(s))
		for i, val := range s {
			ret[i] = FromValue(val)
		}
		return ret
	} else if _, ok := v.(Null); ok {
		return nil
	} else if s, ok := v.(str); ok {
		return s.content
	} else if s, ok := v.(it); ok {
		return s.value
	} else if s, ok := v.(bol); ok {
		return s.value
	} else if s, ok := v.(bs); ok {
		return string(s.content)
	}
	return v.String()
}

type Variable struct {
	Name    string
	Args    []Value
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
		v.value = f(v.Context, v.Args...)
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
	variables   *Complex
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
		variables:   NewComplex(),
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

func (ctx *Context) Declare(names ...string) {
	for _, name := range names {
		ctx.vars[name] = struct{}{}
	}
}

func (ctx *Context) BindValue(key string, val Value) {
	if cctx := ctx.declareAt(key); cctx != nil {
		cctx.bindValue(key, val)
		return
	}
	ctx.bindValue(key, val)
}

func (ctx *Context) bindValue(key string, val Value) {
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

func (ctx *Context) declareAt(name string) *Context {
	idx := strings.Index(name, ".")
	if idx > -1 {
		name = name[:idx]
	}
	if _, ok := ctx.vars[name]; ok {
		return ctx
	}
	if ctx.parent != nil {
		return ctx.parent.declareAt(name)
	}
	return nil
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
