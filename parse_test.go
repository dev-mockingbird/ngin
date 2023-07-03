// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin_test

import (
	"bytes"
	"testing"

	"github.com/dev-mockingbird/ngin"
	"github.com/google/uuid"
)

func TestParse(t *testing.T) {
	bs := bytes.NewBuffer([]byte(`
# the whitespace is very important, please remember use it around any meaningful token
{
	var [ header response host path ];
    header.request-id == null {
        header.request-id = uuid;
    }
    response.header.request-id = header.request-id;
    listen 6000 {
        host == hello.com | world.com {
            backend 127.0.0.1:6090 | 127.0.0.1:6091;
            header.Authorization ~ .+ {
                call [ POST http://127.0.0.1:6080/authentication ];
                response.code == 200 {
                    header.user-id = response.userId;
                }
				forward;
                return;
            }
            path !~ /login | /register | /idinfo/* {
                response.code = 401;
                response.body = "unauthorized";
            }
        }
    }
}
	`))
	lexer := ngin.NewLexer()
	p := ngin.Parser{Lexer: lexer, Reader: bs}
	stmts, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	ctx := ngin.NewContext()
	var listenExecuted, backendExecuted, callExecuted, forwardExecuted bool
	ctx.BindFunc("listen", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		listenExecuted = true
		return true, nil
	})
	ctx.BindFunc("backend", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		backendExecuted = true
		return true, nil
	})
	ctx.BindFunc("call", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		callExecuted = true
		return true, nil
	})
	ctx.BindFunc("forward", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		forwardExecuted = true
		return true, nil
	})
	ctx.BindValue("header.Authorization", ngin.String("xxx"))
	ctx.BindValue("host", ngin.String("hello.com"))
	ctx.BindValuedFunc("uuid", func(ctx *ngin.Context) ngin.Value {
		return ngin.String(uuid.NewString())
	})
	for _, s := range stmts {
		ok, err := s.Execute(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			break
		}
	}
	if !listenExecuted || !backendExecuted || !callExecuted || !forwardExecuted {
		t.Fatal("execute failed")
	}
}
