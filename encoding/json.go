// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package encoding

import (
	"encoding/json"
	"reflect"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func Init(ctx *ngin.Context) {
	ctx.BindValuedFunc("decode-json", DecodeJson)
	ctx.BindValuedFunc("encode-json", EncodeJson)
	ctx.BindValuedFunc("decode-base64", decodeBase64)
	ctx.BindValuedFunc("encode-base64", encodeBase64)
}

func DecodeJson(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide the args for json decode")
		return ngin.Null{}
	}
	ret := make(map[string]any)
	bs := args[0].WithContext(ctx).Bytes()
	if err := json.Unmarshal(bs, &ret); err != nil {
		ctx.Logger().Logf(logf.Error, "json unmarshal: %s", err.Error())
		return ngin.Null{}
	}
	return ngin.ToValue(ret)
}

func EncodeJson(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide the args for json encode")
		return ngin.Null{}
	}
	m := ngin.FromValue(args[0].WithContext(ctx))
	mr := reflect.ValueOf(m)
	bs, err := json.Marshal(mr.Addr().Interface())
	if err != nil {
		ctx.Logger().Logf(logf.Error, "json marshal: %s", err.Error())
		return ngin.Null{}
	}
	return ngin.Bytes(bs)
}
