// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"net/url"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func backend(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	backends := make(map[string]struct{})
	for _, arg := range args {
		for _, v := range arg.Slice() {
			backends[v.String()] = struct{}{}
		}
	}
	for b := range backends {
		u, err := url.Parse(b)
		if err != nil {
			ctx.Logger().Logf(logf.Error, "parse url: %s", err.Error())
			return true, err
		}
		ctx.BindValue("host", ngin.String(u.Host))
		ctx.BindValue("scheme", ngin.String(u.Scheme))
		break
	}
	return true, nil
}
