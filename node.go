package nginz

import (
	"errors"
	"net/http"
)

type Context struct {
	Request  *http.Request
	Response *http.Response
}

func (c *Context) DeepCopyInto(a *Context) {

}

type StmtType int

type Stmt interface {
	Execute(c *Context) (bool, error)
}

type StmtBlock []Stmt

func (stmts StmtBlock) Execute(ctx *Context) (bool, error) {
	for _, stmt := range stmts {
		ok, err := stmt.Execute(ctx)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

type Key []string
type Operation int

type ValueType int

const (
	ValueDirect = iota
	ValueVariable
)

type Value struct {
	Type ValueType
}

type Variable struct {
	Name string
}

func (v Variable) matchName(n string) (bool, string) {
	if len(v.Name) < len(n) {
		return false, ""
	}
	return v.Name[:len(n)] == n && v.Name[len(n)] == '.', v.Name[len(n):]
}

func (v Variable) Value(ctx *Context) (any, error) {
	for name, val := range map[string]any{
		"host": ctx.Request.Host,
		"header": func(sub string) any {
			return ctx.Request.Header.Get(sub)
		},
		"scheme": ctx.Request.URL.Scheme,
		"query": func(sub string) any {
			return ctx.Request.URL.Query().Get(sub)
		},
		"hash":   ctx.Request.URL.Fragment,
		"method": ctx.Request.Method,
		"path":   ctx.Request.URL.Path,
		"response": func(sub string) any {
			return nil
		},
	} {
		ok, sub := v.matchName(name)
		if !ok {
			continue
		}
		if call, ok := val.(func(string) any); ok {
			return call(sub), nil
		}
		return val, nil
	}
	return nil, errors.New("undefined variable")
}

func (v Variable) SetValue(ctx *Context, value any) error {
	for name, val := range map[string]any{
		"host": &ctx.Request.Host,
		"header": func(sub string) {
			ctx.Request.Header.Set(sub, value.(string))
		},
		"scheme": &ctx.Request.URL.Scheme,
		"query": func(sub string) {
			ctx.Request.URL.Query().Set(sub, value.(string))
		},
		"path":   &ctx.Request.URL.Path,
		"method": &ctx.Request.Method,
		"response": func(sub string) {

		},
	} {
		ok, sub := v.matchName(name)
		if !ok {
			continue
		}
		if call, ok := val.(func(string)); ok {
			call(sub)
			return nil
		} else if str, ok := val.(*string); ok {
			*str = value.(string)
		} else if i, ok := val.(*int); ok {
			*i = val.(int)
		}
		return errors.New("can't set value")
	}
	return nil
}

const (
	EQ = iota
	Like
	NEQ
	In
	NotIn
	GT
	LT
	GTE
	LTE
)

type MatchStmt struct {
	Key       Key
	Operation Operation
	Value     any
}

func (stmt MatchStmt) Execute(req *http.Request, res *http.Response) (bool, error) {

}

func (stmt MatchStmt) leftValue(req *http.Request, res *http.Response) (any, error) {

}
