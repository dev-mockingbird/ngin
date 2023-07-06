// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin

import (
	"errors"
	"fmt"
	"io"
)

func ErrUnexpectedToken(t *Token) error {
	return fmt.Errorf("unexpected token [%s] at [%d, %d]", t.String(), t.Row, t.Col)
}

var operatorMap map[int]Operator

func init() {
	operatorMap = map[int]Operator{
		TokenEQ:      EQ,
		TokenNEQ:     NEQ,
		TokenGTE:     GTE,
		TokenGT:      GT,
		TokenLTE:     LTE,
		TokenLike:    Like,
		TokenNotLike: NotLike,
	}
}

// Stmt -> Stmt? { Stmt;* } | BoolStmt; | AssignmentStmt; | CallStmt;
// BoolStmt -> ComparableValue Operator ComparableValue
// ComparableValue -> Array | Name | Value
// Array -> Name | Value '|' Name | Value
// Operator -> GT | LT | GTE | LTE | LIKE | NotLike | EQ | NEQ
// AssignmentStmt -> Name = Value
// CallStmt -> Name [ Name | Value *]
type Parser struct {
	Lexer  *Lexer
	Reader io.Reader
	token  Token
}

func (p *Parser) Parse() (ret []Stmt, err error) {
	for {
		if err = p.nextToken(); err != nil {
			return
		}
		if p.token.Type == TokenEOF {
			p.useToken()
			return
		}
		var stmt Stmt
		if stmt, err = p.Stmt(); err != nil {
			return
		}
		ret = append(ret, stmt)
	}
}

func (p *Parser) Stmt() (Stmt, error) {
	switch p.token.Type {
	case TokenBlockBegin:
		stmt := MatchThenStmt{Match: EmptyStmt{}, Stmts: []Stmt{}}
		p.useToken()
		for {
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			s, err := p.Stmt()
			if err != nil {
				return nil, err
			}
			if _, ok := s.(EmptyStmt); ok {
				return stmt, nil
			}
			stmt.Stmts = append(stmt.Stmts, s)
		}
	case TokenBlockEnd:
		p.useToken()
		return EmptyStmt{}, nil
	default:
		stmt, err := p.smallStmt()
		if err != nil {
			return stmt, err
		}
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if p.token.Type == TokenStmtEnd {
			p.useToken()
			return stmt, nil
		} else if p.token.Type != TokenBlockBegin {
			return nil, ErrUnexpectedToken(&p.token)
		}
		matchThenStmt, err := p.Stmt()
		if err != nil {
			return nil, err
		}
		mts, ok := matchThenStmt.(MatchThenStmt)
		if !ok {
			// this should never reach
			return nil, errors.New("this should never be reached")
		}
		mts.Match = stmt
		return mts, nil
	}
}

func (p *Parser) smallStmt() (Stmt, error) {
	switch p.token.Type {
	case TokenReturn:
		p.useToken()
		return ReturnStmt{}, nil
	default:
		v, err := p.nameOrValue()
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, ErrUnexpectedToken(&p.token)
		}
		var right Value
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		if variable, ok := v.(*Variable); ok {
			switch p.token.Type {
			case TokenAssignment:
				p.useToken()
				if err := p.nextToken(); err != nil {
					return nil, err
				}
				if right, err = p.nameOrValue(); err != nil {
					return nil, err
				}
				return AssignmentStmt{Name: variable.Name, Value: right}, nil
			case TokenStmtEnd, TokenBlockBegin:
				return FuncStmt{Name: variable.Name, Args: variable.Args}, nil
			}
		}
		var ok bool
		var operator Operator
		if operator, ok = operatorMap[p.token.Type]; ok {
			p.useToken()
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			if right, err = p.nameOrValue(); err != nil {
				return nil, err
			}
			return MatchStmt{Left: v, Operator: operator, Right: right}, nil
		}
		return nil, ErrUnexpectedToken(&p.token)
	}
}

func (p *Parser) nameOrValue() (Value, error) {
	getValue := func() Value {
		switch p.token.Type {
		case TokenString, TokenInt, TokenFloat:
			return Bytes(p.token.Raw)
		case TokenName:
			return &Variable{Name: string(p.token.Raw)}
		case TokenFalse:
			return Bool(false)
		case TokenTrue:
			return Bool(true)
		case TokenNull:
			return Null{}
		default:
			return nil
		}
	}
	ret := getValue()
	if ret == nil {
		return ret, nil
	}
	p.useToken()
	va, isVar := ret.(*Variable)
	for {
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		switch {
		case p.token.Type == TokenSep:
			p.useToken()
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			n := getValue()
			if isVar && len(va.Args) > 0 {
				p.useToken()
				v := va.Args[len(va.Args)-1]
				if v != nil {
					r := v.Slice()
					r = append(r, n)
					va.Args[len(va.Args)-1] = Slice(r)
					continue
				}
			} else if !isVar {
				p.useToken()
				r := ret.Slice()
				r = append(r, n)
				ret = Slice(r)
				continue
			} else {
				ret = Slice([]Value{ret, n})
				p.useToken()
				isVar = false
				continue
			}
			return ret, ErrUnexpectedToken(&p.token)
		case isVar:
			v := getValue()
			if v == nil {
				return va, nil
			}
			va.Args = append(va.Args, v)
			p.useToken()
		default:
			return ret, nil
		}
	}
}

func (p *Parser) nextToken() (err error) {
	if p.token.Type != TokenEmpty {
		return nil
	}
	if p.token, err = p.Lexer.Scan(p.Reader); err != nil {
		return err
	}
	if p.token.Type == TokenComment {
		p.useToken()
		return p.nextToken()
	}
	// fmt.Printf("token: %s\n", p.token.String())
	return
}

func (p *Parser) useToken(use ...func()) {
	if len(use) > 0 {
		use[0]()
	}
	p.token = Token{Type: TokenEmpty}
}
