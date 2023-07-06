// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package log

import (
	"strings"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func Init(ctx *ngin.Context) {
	ctx.BindFunc("log", Log)
}

func Log(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	if len(args) < 1 {
		ctx.Logger().Logf(logf.Warn, "log with nothing")
		return true, nil
	} else if len(args) < 2 {
		ctx.Logger().Logf(logf.Info, args[0].String())
		return true, nil
	}
	levels := map[string]logf.Level{
		"trace": logf.Trace,
		"debug": logf.Debug,
		"info":  logf.Info,
		"warn":  logf.Warn,
		"error": logf.Error,
		"fatal": logf.Fatal,
	}
	level, ok := levels[strings.ToLower(args[0].String())]
	if !ok {
		ctx.Logger().Logf(logf.Warn, "unkown log level: %s", args[0].String())
		level = logf.Info
	}
	ctx.Logger().Logf(level, args[1].String(), func() []any {
		ret := []any{}
		for i := 2; i < len(args); i++ {
			ret = append(ret, args[i].String())
		}
		return ret
	}()...)
	return true, nil
}
