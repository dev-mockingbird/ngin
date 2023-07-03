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
	var left, right Value
	var operator Operator
	switch p.token.Type {
	case TokenReturn:
		p.useToken()
		return ReturnStmt{}, nil
	case TokenName:
		left = &Variable{Name: string(p.token.Raw)}
		p.useToken()
		if err := p.nextToken(); err != nil {
			return nil, err
		}
		var ok bool
		var err error
		if operator, ok = operatorMap[p.token.Type]; ok {
			p.useToken()
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			if right, err = p.nameOrValue(); err != nil {
				return nil, err
			}
			return MatchStmt{Left: left, Operator: operator, Right: right}, nil
		} else if p.token.Type == TokenAssignment {
			p.useToken()
			if err := p.nextToken(); err != nil {
				return nil, err
			}
			if right, err = p.nameOrValue(); err != nil {
				return nil, err
			}
			return AssignmentStmt{Name: left.(*Variable).Name, Value: right}, nil
		} else if p.token.Type == TokenFuncArgBegin {
			p.useToken()
			var args []Value
			for {
				if err = p.nextToken(); err != nil {
					return nil, err
				}
				var v Value
				if v, err = p.nameOrValue(); err != nil {
					return nil, err
				} else if v == nil {
					if p.token.Type == TokenFuncArgEnd {
						p.useToken()
						return FuncStmt{Name: left.(*Variable).Name, Args: args}, nil
					}
					return nil, ErrUnexpectedToken(&p.token)
				}
				args = append(args, v)
			}
		}
		var v Value
		if v, err = p.nameOrValue(); err != nil {
			return nil, err
		} else if v == nil {
			// no args func
			return FuncStmt{Name: left.(*Variable).Name}, nil
		}
		return FuncStmt{Name: left.(*Variable).Name, Args: []Value{v}}, nil
	default:
		var err error
		var ok bool
		// as a match statement
		if left, err = p.nameOrValue(); err != nil {
			return nil, err
		}
		if left == nil {
			return nil, ErrUnexpectedToken(&p.token)
		}
		p.useToken()
		if operator, ok = operatorMap[p.token.Type]; !ok {
			return nil, ErrUnexpectedToken(&p.token)
		}
		p.useToken()
		if right, err = p.nameOrValue(); err != nil {
			return nil, err
		}
		if right == nil {
			return nil, ErrUnexpectedToken(&p.token)
		}
		return MatchStmt{Left: left, Operator: operator, Right: right}, nil
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
			if v := getValue(); v != nil {
				p.useToken()
				r := ret.Slice()
				r = append(r, v)
				ret = Slice(r)
				continue
			}
			return ret, errors.New("unexpected token")
		}
		return ret, nil
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
