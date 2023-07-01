package ngin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const (
	TokenEmpty        = iota
	TokenNull         // 'null'
	TokenReturn       // 'return'
	TokenTrue         // 'true'
	TokenFalse        // 'false'
	TokenEQ           // '=='
	TokenNEQ          // '!='
	TokenGTE          // '>='
	TokenGT           // '>'
	TokenLTE          // '<='
	TokenLT           // '<'
	TokenAssignment   // '='
	TokenOR           // '||'
	TokenAND          // '&&'
	TokenSep          // '|'
	TokenLike         // '~'
	TokenNotLike      // '!~'
	TokenStrSep       // '"'
	TokenStmtEnd      // ';'
	TokenBlockBegin   // '{'
	TokenBlockEnd     // '}'
	TokenFuncArgBegin // '['
	TokenFuncArgEnd   // ']'
	TokenName         // ''
	TokenFloat        // ''
	TokenComment      // '# xxxx\n'
	TokenInt
	TokenBool
	TokenString
	TokenEOF
)

var tokenMap map[int]string

func init() {
	tokenMap = map[int]string{
		TokenNull:         "null",
		TokenReturn:       "return",
		TokenTrue:         "true",
		TokenFalse:        "false",
		TokenEQ:           "==",
		TokenNEQ:          "!=",
		TokenGTE:          ">=",
		TokenGT:           ">",
		TokenLTE:          "<=",
		TokenLT:           "<",
		TokenAssignment:   "=",
		TokenOR:           "||",
		TokenAND:          "&&",
		TokenSep:          "|",
		TokenLike:         "~",
		TokenNotLike:      "!~",
		TokenStmtEnd:      ";",
		TokenBlockBegin:   "{",
		TokenBlockEnd:     "}",
		TokenFuncArgBegin: "[",
		TokenFuncArgEnd:   "]",
	}
}

const (
	stateStart = iota
	stateNull
	stateReturn
	stateEqual
	stateGTE
	stateGT
	stateLTE
	stateLT
	stateAssignment
	stateComment
	stateNot
	stateName
	stateNumber
	stateFloat
	stateString
	stateEnd
	stateTrue
	stateFalse
)

var (
	null = []byte{'n', 'u', 'l', 'l'}
	fls  = []byte{'f', 'a', 'l', 's', 'e'}
	tru  = []byte{'t', 'r', 'u', 'e'}
	rtrn = []byte{'r', 'e', 't', 'u', 'r', 'n'}
)

type UnexpectedChar byte

type PosError struct {
	Row int
	Col int
	err error
}

func (e PosError) Error() string {
	return fmt.Sprintf("%s at %d, %d", e.err.Error(), e.Row, e.Col)
}

func (c UnexpectedChar) Error() string {
	return fmt.Sprintf("unexpected char '%s'", []byte{byte(c)})
}

type Token struct {
	Type int
	bs   []byte
}

func (t Token) String() string {
	if s, ok := tokenMap[t.Type]; ok {
		return s
	}
	return string(t.bs)
}

type Lexer struct {
	b     []byte
	state int
	stash []byte
	col   int
	row   int
}

func NewLexer() *Lexer {
	return &Lexer{b: make([]byte, 1)}
}

func (l *Lexer) Scan(r io.Reader) (t Token, err error) {
	if l.b == nil {
		l.b = make([]byte, 1)
	}
	for {
		if len(l.stash) > 0 {
			l.b[0] = l.stash[0]
			l.stash = l.stash[1:]
		} else {
			_, err = r.Read(l.b)
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
					t.Type = TokenEOF
					return
				}
				return
			}
		}
		switch l.state {
		case stateStart:
			err = l.stateStart(&t)
		case stateNull:
			err = l.stateNull(&t)
		case stateReturn:
			err = l.stateReturn(&t)
		case stateFalse:
			err = l.stateFalse(&t)
		case stateTrue:
			err = l.stateTrue(&t)
		case stateComment:
			err = l.stateComment(&t)
		case stateName:
			err = l.stateName(&t)
		case stateString:
			err = l.stateString(&t)
		case stateNumber:
			err = l.stateNumber(&t)
		case stateFloat:
			err = l.stateFloat(&t)
		case stateNot:
			err = l.stateNot(&t)
		case stateEqual:
			err = l.stateEqual(&t)
		case stateAssignment:
			err = l.stateAssignment(&t)
		case stateGT:
			err = l.stateGT(&t)
		case stateGTE:
			err = l.stateGTE(&t)
		case stateLT:
			err = l.stateLT(&t)
		case stateLTE:
			err = l.stateLTE(&t)
		}
		if l.state == stateEnd {
			l.state = stateStart
			return
		}
		if err != nil {
			if len(l.stash) == 0 {
				err = PosError{err: err, Col: l.col, Row: l.row}
				return
			}
			col := l.col
			row := l.row - bytes.Count(l.stash, []byte{'\n'})
			if row != l.row {
				col -= len(l.stash[bytes.LastIndex(l.stash, []byte{'\n'}):])
			}
			err = PosError{err: err, Col: col, Row: row}
			return
		}
		if l.b[0] == '\n' {
			l.col = 0
			l.row++
			continue
		}
		l.col++
	}
}

func (l *Lexer) stateNumber(t *Token) error {
	switch {
	case l.b[0] == '.':
		t.bs = append(t.bs, '.')
		l.state = stateFloat
	case l.isNumber():
		t.bs = append(t.bs, l.b[0])
	case l.isWhitespace():
		t.Type = TokenInt
		l.state = stateEnd
	case l.isSep():
		t.Type = TokenInt
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
	default:
		l.state = stateString
		t.bs = append(t.bs, l.b[0])
	}
	return nil
}

func (l *Lexer) stateFloat(t *Token) error {
	switch {
	case l.isNumber():
		t.bs = append(t.bs, l.b[0])
	case l.isWhitespace():
		t.Type = TokenFloat
		l.state = stateEnd
	case l.isSep():
		t.Type = TokenFloat
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
	default:
		l.state = stateString
		t.bs = append(t.bs, l.b[0])
	}
	return nil
}

func (l *Lexer) stateComment(t *Token) error {
	if l.b[0] == '\n' {
		t.Type = TokenComment
		l.state = stateEnd
		return nil
	}
	t.bs = append(t.bs, l.b[0])
	return nil
}

func (l *Lexer) stateEqual(t *Token) error {
	switch {
	case l.isWhitespace():
		t.Type = TokenEQ
		l.state = stateEnd
	default:
		l.state = stateString
		l.stash = []byte{l.b[0]}
	}
	return nil
}

func (l *Lexer) stateNull(t *Token) error {
	switch {
	case len(t.bs) < 4 && null[len(t.bs)] == l.b[0]:
		t.bs = append(t.bs, l.b[0])
		return nil
	case l.isWhitespace():
		t.Type = TokenNull
		l.state = stateEnd
		return nil
	case l.isSep():
		t.Type = TokenNull
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
		return nil
	default:
		l.state = stateName
		return nil
	}
}

func (l *Lexer) stateTrue(t *Token) error {
	switch {
	case len(t.bs) < 4 && tru[len(t.bs)] == l.b[0]:
		t.bs = append(t.bs, l.b[0])
		return nil
	case l.isWhitespace():
		t.Type = TokenTrue
		l.state = stateEnd
		return nil
	case l.isSep():
		t.Type = TokenTrue
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
		return nil
	default:
		l.state = stateName
		return nil
	}
}

func (l *Lexer) stateFalse(t *Token) error {
	switch {
	case len(t.bs) < 4 && fls[len(t.bs)] == l.b[0]:
		t.bs = append(t.bs, l.b[0])
		return nil
	case l.isWhitespace():
		t.Type = TokenFalse
		l.state = stateEnd
		return nil
	case l.isSep():
		t.Type = TokenFalse
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
		return nil
	default:
		l.state = stateName
		return nil
	}
}

func (l *Lexer) stateReturn(t *Token) error {
	switch {
	case len(t.bs) < 6 && rtrn[len(t.bs)] == l.b[0]:
		t.bs = append(t.bs, l.b[0])
	case l.isWhitespace():
		t.Type = TokenReturn
		l.state = stateEnd
	case l.isSep():
		t.Type = TokenReturn
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
	default:
		l.state = stateName
		return nil
	}
	return nil
}

func (l *Lexer) stateNot(t *Token) error {
	if l.isWhitespace() {
		return nil
	}
	switch l.b[0] {
	case '=':
		t.Type = TokenNEQ
		l.state = stateEnd
	case '~':
		t.Type = TokenNotLike
		l.state = stateEnd
	default:
		t.Type = TokenString
		t.bs = []byte{'!', l.b[0]}
	}
	return nil
}

func (l *Lexer) stateAssignment(t *Token) error {
	if l.isWhitespace() {
		t.Type = TokenAssignment
		return nil
	}
	switch {
	case l.b[0] == '=':
		l.state = stateEqual
	default:
		t.Type = TokenAssignment
		l.stash = []byte{l.b[0]}
		l.state = stateEnd
	}
	return nil
}

func (l *Lexer) stateName(t *Token) error {
	switch {
	case l.isName():
		t.bs = append(t.bs, l.b[0])
		return nil
	case l.isWhitespace():
		t.Type = TokenName
		l.state = stateEnd
	case l.isSep():
		t.Type = TokenName
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
	default:
		l.state = stateString
		t.bs = append(t.bs, l.b[0])
	}
	return nil
}

func (l *Lexer) stateString(t *Token) error {
	switch {
	case l.b[0] == '"' && t.bs[0] == '"':
		t.Type = TokenString
		t.bs = t.bs[1:]
		l.state = stateEnd
	case l.b[0] == ';':
		t.Type = TokenString
		l.state = stateEnd
		l.stash = []byte{l.b[0]}
	case l.isWhitespace():
		t.Type = TokenString
		l.state = stateEnd
	default:
		t.bs = append(t.bs, l.b[0])
	}
	return nil
}

func (l *Lexer) stateLTE(t *Token) error {
	switch {
	case l.isWhitespace():
		t.Type = TokenGTE
		l.state = stateEnd
	default:
		l.state = stateString
		t.bs = []byte{'<', '=', l.b[0]}
	}
	return nil
}

func (l *Lexer) stateLT(t *Token) error {
	switch {
	case l.b[0] == '=':
		l.state = stateLTE
	case l.isWhitespace():
		l.state = stateEnd
		t.Type = TokenLT
	default:
		l.state = stateString
		t.bs = []byte{'<', l.b[0]}
	}
	return nil
}

func (l *Lexer) stateGTE(t *Token) error {
	switch {
	case l.isWhitespace():
		t.Type = TokenGTE
		l.state = stateEnd
	default:
		l.state = stateString
		t.bs = []byte{'>', '=', l.b[0]}
	}
	return nil
}

func (l *Lexer) stateGT(t *Token) error {
	switch {
	case l.b[0] == '=':
		l.state = stateGTE
	case l.isWhitespace():
		t.Type = TokenGT
		l.state = stateEnd
	default:
		l.state = stateString
		t.bs = append(t.bs, l.b[0])
	}
	return nil
}

func (l *Lexer) stateStart(t *Token) error {
	switch l.b[0] {
	case 't':
		t.bs = append(t.bs, l.b[0])
		l.state = stateTrue
	case 'f':
		t.bs = append(t.bs, l.b[0])
		l.state = stateFalse
	case 'n':
		t.bs = append(t.bs, l.b[0])
		l.state = stateNull
	case 'r':
		t.bs = append(t.bs, l.b[0])
		l.state = stateReturn
	case '{':
		t.Type = TokenBlockBegin
		l.state = stateEnd
	case '}':
		t.Type = TokenBlockEnd
		l.state = stateEnd
	case '[':
		t.Type = TokenFuncArgBegin
		l.state = stateEnd
	case ']':
		t.Type = TokenFuncArgEnd
		l.state = stateEnd
	case '!':
		l.state = stateNot
	case '=':
		l.state = stateAssignment
	case '>':
		l.state = stateGT
	case '<':
		l.state = stateLT
	case '~':
		t.Type = TokenLike
		l.state = stateEnd
	case '|':
		t.Type = TokenSep
		l.state = stateEnd
	case ';':
		t.Type = TokenStmtEnd
		l.state = stateEnd
	case '#':
		l.state = stateComment
	case '"':
		l.state = stateString
		t.bs = append(t.bs, '"')
	default:
		switch {
		case l.isAlpha() || l.b[0] == '_' || l.b[0] == '-':
			t.bs = append(t.bs, l.b[0])
			l.state = stateName
		case l.isNumber():
			t.bs = append(t.bs, l.b[0])
			l.state = stateNumber
		case l.isWhitespace():
			// auto clean whitespace
		default:
			t.bs = append(t.bs, l.b[0])
			l.state = stateString
		}
	}
	return nil
}

func (l *Lexer) isWhitespace() bool {
	return l.b[0] == '\t' || l.b[0] == ' ' || l.b[0] == '\n'
}

func (l *Lexer) isAlpha() bool {
	return l.b[0] >= 'a' && l.b[0] <= 'z' || l.b[0] >= 'A' && l.b[0] <= 'Z'
}

func (l *Lexer) isNumber() bool {
	return l.b[0] >= '0' && l.b[0] <= '9'
}

func (l *Lexer) isSep() bool {
	return l.b[0] == ';'
}

func (l *Lexer) isName() bool {
	return l.isAlpha() || l.isNumber() || l.b[0] == '_' || l.b[0] == '-' || l.b[0] == '.'
}
