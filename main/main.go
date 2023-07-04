// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/dev-mockingbird/ngin"
)

func main() {
	var confPath string
	flag.StringVar(&confPath, "config", "/etc/ngin/config.ngin", "pathfile of the config")
	flag.Parse()
	fs, err := os.OpenFile(confPath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("open config file: %s\n", err.Error())
		os.Exit(1)
	}
	ctx := ngin.NewContext()
	ctx.BindFunc("listen", listen)
	ctx.BindFunc("call", call)
	ctx.BindFunc("backend", backend)
	ctx.BindValuedFunc("decode-json", decodeJson)
	ctx.BindValuedFunc("encode-json", encodeJson)
	ctx.BindFunc("redis-cfg", redis_cfg)
	ctx.BindFunc("redis-set", redis_set)
	ctx.BindValuedFunc("redis-get", redis_get)
	parser := ngin.Parser{Lexer: ngin.NewLexer(), Reader: fs}
	stmts, err := parser.Parse()
	if err != nil {
		fmt.Printf("parse: %s\n", err.Error())
		os.Exit(1)
	}
	var wg sync.WaitGroup
	for _, stmt := range stmts {
		wg.Add(1)
		go func(stmt ngin.Stmt, ctx *ngin.Context) {
			ctx = ctx.Folk()
			if _, err := stmt.Execute(ctx); err != nil {
				fmt.Printf("%s\n", err.Error())
			}
			wg.Done()
		}(stmt, ctx)
	}
	wg.Wait()
}
