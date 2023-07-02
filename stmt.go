package ngin

import (
	"errors"
	"fmt"
	"regexp"
)

type Stmt interface {
	Execute(ctx *Context) (bool, error)
}

type ReturnStmt struct{}

func (ReturnStmt) Execute(ctx *Context) (bool, error) {
	return false, nil
}

type Operator int

const (
	EQ = iota
	NEQ
	GTE
	GT
	LTE
	LT
	Like
	NotLike
)

type ValueAble interface {
	Value() Value
}

type MatchStmt struct {
	Left, Right Value
	Operator    Operator
}

func (m MatchStmt) Execute(ctx *Context) (bool, error) {
	left := m.Left
	if v, ok := left.(*Variable); ok {
		v.Context = ctx
		left = v
	}
	right := m.Right
	if v, ok := right.(*Variable); ok {
		v.Context = ctx
		right = v
	}
	switch m.Operator {
	case EQ:
		if s, ok := right.(Slice); ok {
			return s.Contain(left), nil
		}
		r := left.Compare(right)
		return r == 0, nil
	case NEQ:
		if s, ok := right.(Slice); ok {
			return !s.Contain(left), nil
		}
		r := left.Compare(right)
		return r != 0, nil
	case GTE:
		r := left.Compare(right)
		return r >= 0, nil
	case GT:
		r := left.Compare(right)
		return r > 0, nil
	case LTE:
		r := left.Compare(right)
		return r <= 0, nil
	case LT:
		r := left.Compare(right)
		return r < 0, nil
	case Like:
		like := func(l, r string) (bool, error) {
			re, err := regexp.Compile(r)
			if err != nil {
				return false, err
			}
			return re.MatchString(l), nil
		}
		if s, ok := right.(Slice); ok {
			for _, item := range s {
				if ok, err := like(left.String(), item.String()); err != nil || ok {
					return ok, err
				}
			}
			return false, nil
		}
		return like(left.String(), right.String())
	case NotLike:
		notlike := func(l, r string) (bool, error) {
			re, err := regexp.Compile(r)
			if err != nil {
				return false, err
			}
			return !re.MatchString(l), nil
		}
		if s, ok := right.(Slice); ok {
			for _, item := range s {
				if ok, err := notlike(left.String(), item.String()); err != nil || ok {
					return ok, err
				}
			}
			return false, nil
		}
		return notlike(left.String(), right.String())
	default:
		return false, errors.New("not supported operator")
	}
}

type AssignmentStmt struct {
	Name  string
	Value Value
}

func (a AssignmentStmt) Execute(ctx *Context) (bool, error) {
	if v, ok := a.Value.(*Variable); ok {
		v.Context = ctx
	}
	ctx.RegisterVariable(a.Name, a.Value)
	return true, nil
}

type FuncStmt struct {
	Name string
	Args []Value
}

func (f FuncStmt) Execute(ctx *Context) (bool, error) {
	if funk, ok := ctx.funks[f.Name]; ok {
		return funk(ctx, f.Args...)
	}
	return false, fmt.Errorf("func [%s] not found", f.Name)
}

type EmptyStmt struct{}

func (EmptyStmt) Execute(ctx *Context) (bool, error) {
	return true, nil
}

type MatchThenStmt struct {
	Match Stmt
	Stmts []Stmt
}

func (mt MatchThenStmt) Execute(ctx *Context) (bool, error) {
	ctx.stmts = mt.Stmts
	matched, err := mt.Match.Execute(ctx)
	if err != nil || !matched {
		return matched, err
	}
	ctx.stmts = nil
	for _, s := range mt.Stmts {
		con, err := s.Execute(ctx)
		if !con || err != nil {
			return con, err
		}
	}
	return true, nil
}
