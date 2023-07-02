// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dev-mockingbir/ngin"
	"github.com/google/uuid"
)

func TestParse(t *testing.T) {
	bs := bytes.NewBuffer([]byte(`
# the whitespace is very important, please remember use it around any meaningful token
{
    header.request-id == null {
        header.request-id = uuid;
    }
    response.header.request-id = header.request-id;
    listen 6000 {
        host == hello.com | world.com {
            backend 127.0.0.1:6090 | 127.0.0.1:6091;
            header.Authorization ~ * {
                call [ "POST" http://127.0.0.1:6080/authentication ];
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
	ctx.BindFunc("listen", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		fmt.Printf("listened %s\n", args[0].String())
		return true, nil
	})
	ctx.BindFunc("backend", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		args = args[0].Slice()
		fmt.Printf("packend [%s %s] set\n", args[0].String(), args[1].String())
		return true, nil
	})
	ctx.BindFunc("call", func(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
		fmt.Printf("call [%s %s] set\n", args[0].String(), args[1].String())
		return true, nil
	})
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
}
