// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package ngin_test

import (
	"bytes"
	"testing"

	"github.com/dev-mockingbir/ngin"
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
                call [ POST http://127.0.0.1:6080/authentication ];
                response.code == 200 {
                    header.user-id = response.userId;
                }
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
	t.Logf("%#v\n", stmts)
}
