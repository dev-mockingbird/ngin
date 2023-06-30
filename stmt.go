package ngin

import (
	"errors"
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
	In
	NotIn
	Like
)

type ValueAble interface {
	Value() Value
}

type MatchStmt struct {
	Left, Right ValueAble
	Operator    Operator
}

func (m MatchStmt) Execute(ctx *Context) (bool, error) {
	switch m.Operator {
	case EQ:
		r := m.Left.Value().Compare(m.Right.Value())
		return r == 0, nil
	case NEQ:
		r := m.Left.Value().Compare(m.Right.Value())
		return r != 0, nil
	case GTE:
		r := m.Left.Value().Compare(m.Right.Value())
		return r >= 0, nil
	case GT:
		r := m.Left.Value().Compare(m.Right.Value())
		return r > 0, nil
	case LTE:
		r := m.Left.Value().Compare(m.Right.Value())
		return r <= 0, nil
	case LT:
		r := m.Left.Value().Compare(m.Right.Value())
		return r < 0, nil
	case In:
		for _, v := range m.Right.Value().Slice() {
			if m.Left.Value().Compare(v) == 0 {
				return true, nil
			}
		}
		return false, nil
	case NotIn:
		for _, v := range m.Right.Value().Slice() {
			if m.Left.Value().Compare(v) == 0 {
				return false, nil
			}
		}
		return true, nil
	case Like:
		r := m.Right.Value().String()
		re, err := regexp.Compile(r)
		if err != nil {
			return false, err
		}
		return re.Match(m.Left.Value().Bytes()), nil
	default:
		return false, errors.New("not supported operator")
	}
}

type AssignmentStmt struct {
	Name  string
	Value Value
}

func (a AssignmentStmt) Execute(ctx *Context) (bool, error) {
	ctx.RegisterVariable(a.Name, a.Value)
	return true, nil
}

type FuncStmt struct {
	Name string
	Args []Value
}

func (f FuncStmt) Execute(ctx *Context) (bool, error) {
	if funk, ok := ctx.funks[f.Name]; ok {
		return funk(f.Args...)
	}
	return false, errors.New("func not found")
}

type EmptyStmt struct{}

func (EmptyStmt) Execute(ctx *Context) (bool, error) {
	return true, nil
}

type MatchThenStmt struct {
	Match MatchStmt
	Stmts []Stmt
}

func (mt MatchThenStmt) Execute(ctx *Context) (bool, error) {
	matched, err := mt.Match.Execute(ctx)
	if err != nil || !matched {
		return matched, err
	}
	for _, s := range mt.Stmts {
		con, err := s.Execute(ctx)
		if !con || err != nil {
			return con, err
		}
	}
	return true, nil
}
