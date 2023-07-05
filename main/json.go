// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"reflect"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func decodeJson(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide the args for json decode")
		return ngin.Null{}
	}
	ret := make(map[string]any)
	if err := json.Unmarshal(args[0].Bytes(), &ret); err != nil {
		ctx.Logger().Logf(logf.Error, "json unmarshal: %s", err.Error())
		return ngin.Null{}
	}
	return ngin.ToValue(ret)
}

func encodeJson(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide the args for json encode")
		return ngin.Null{}
	}
	m := ngin.FromValue(args[0])
	mr := reflect.ValueOf(m)
	bs, err := json.Marshal(mr.Addr().Interface())
	if err != nil {
		ctx.Logger().Logf(logf.Error, "json marshal: %s", err.Error())
		return ngin.Null{}
	}
	return ngin.Bytes(bs)
}
