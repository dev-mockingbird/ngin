package ngin

import (
	"bytes"
	"fmt"
	"testing"
)

func TestLex(t *testing.T) {
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
	lexer := NewLexer()
	for {
		token, err := lexer.Scan(bs)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%d, %s\n", token.Type, token.String())
		if token.Type == TokenEOF {
			break
		}
	}
}
