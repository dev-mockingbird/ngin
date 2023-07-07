// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package encoding

import (
	b64 "encoding/base64"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func encodeBase64(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide at least one argument for encode base64")
		return ngin.Null{}
	}
	return ngin.String(b64.URLEncoding.EncodeToString(args[0].WithContext(ctx).Bytes()))
}

func decodeBase64(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide at least one argument for decode base64")
		return ngin.Null{}
	}
	res, err := b64.StdEncoding.DecodeString(args[0].WithContext(ctx).String())
	if err != nil {
		ctx.Logger().Logf(logf.Error, "encode base 64: %s", err.Error())
	}
	return ngin.Bytes(res)
}
