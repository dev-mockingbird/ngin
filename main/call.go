// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"io"
	"net/http"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func requestFromContext(ctx *ngin.Context) (*http.Request, error) {
	r := ctx.Get("request")
	if r == nil {
		ctx.Logger().Logf(logf.Error, "can't get request from context")
		return nil, errors.New("can't get request from context")
	}
	req := r.(*http.Request)
	req.URL.Scheme = ctx.GetValue("scheme").String()
	req.URL.Host = ctx.GetValue("host").String()
	keys := ctx.GetValue("header.*").Slice()
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	for _, key := range keys {
		if k := key.String(); k != "" {
			v := ctx.GetValue("header." + k).String()
			req.Header.Set(k, v)
		}
	}
	keys = ctx.GetValue("query.*").Slice()
	query := req.URL.Query()
	for _, key := range keys {
		if k := key.String(); k != "" {
			query.Set(k, ctx.GetValue("query."+k).String())
		}
	}
	req.URL.RawQuery = query.Encode()
	req.URL.Fragment = ctx.GetValue("hash").String()
	return req, nil
}

func call(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	req, err := requestFromContext(ctx)
	if err != nil {
		return false, err
	}
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		ctx.BindValue("response.code", ngin.String("502"))
		ctx.BindValue("response.body", ngin.String("can't request from backend: "+err.Error()))
		ctx.Logger().Logf(logf.Error, "call: %s", err.Error())
		return false, nil
	}
	for k := range resp.Header {
		ctx.BindValue("response."+k, ngin.String(resp.Header.Get(k)))
	}
	ctx.BindValue("response.code", ngin.Int(uint64(resp.StatusCode)))
	ctx.BindValuedFunc("responseBody", func(_ *ngin.Context, args ...ngin.Value) ngin.Value {
		bs, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			ctx.Logger().Logf(logf.Error, "read response body: %s", err.Error())
			return ngin.Null{}
		}
		return ngin.Bytes(bs)
	})
	return true, nil
}
