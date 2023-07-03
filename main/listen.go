// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"io"
	"net/http"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func withRequest(ctx *ngin.Context, req *http.Request) *ngin.Context {
	ctx.Put("request", req)
	ctx.BindFunc("log", log)
	ctx.BindFunc("listen", listen)
	ctx.BindValuedFunc("body", requestBody)
	for k := range req.Header {
		vals := req.Header.Values(k)
		if len(vals) == 1 {
			ctx.BindValue("header."+k, ngin.String(vals[0]))
			continue
		}
		val := make(ngin.Slice, len(vals))
		for i, v := range vals {
			val[i] = ngin.String(v)
		}
		ctx.BindValue("header."+k, val)
	}
	ctx.BindValue("path", ngin.String(req.URL.Path))
	ctx.BindValue("hash", ngin.String(req.URL.Fragment))
	ctx.BindValue("scheme", ngin.String(req.URL.Scheme))
	ctx.BindValue("host", ngin.String(req.Host))
	ctx.BindValue("port", ngin.String(req.URL.Port()))
	ctx.BindValue("user-agent", ngin.String(req.UserAgent()))
	ctx.BindValue("remote-addr", ngin.String(req.RemoteAddr))
	ctx.BindValue("method", ngin.String(req.Method))
	for k, vals := range req.URL.Query() {
		if len(vals) == 1 {
			ctx.BindValue("query."+k, ngin.String(vals[0]))
			continue
		}
		val := make(ngin.Slice, len(vals))
		for i, v := range vals {
			val[i] = ngin.String(v)
		}
		ctx.BindValue("query."+k, val)
	}
	return ctx
}

func requestBody(ctx *ngin.Context) ngin.Value {
	req := ctx.Get("request")
	if req == nil {
		ctx.Logger().Logf(logf.Warn, "http request is nil")
		return ngin.Null{}
	}
	b, err := io.ReadAll(req.(*http.Request).Body)
	if err != nil {
		ctx.Logger().Logf(logf.Error, "read request body: %s", err.Error())
		return ngin.Null{}
	}
	if err := req.(*http.Request).Body.Close(); err != nil {
		ctx.Logger().Logf(logf.Error, "close body: %s", err.Error())
	}
	return ngin.Bytes(b)
}

func listen(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Fatal, "you shold provide a port after listen")
		return false, nil
	}
	port := ":" + args[0].String()
	certFile := ctx.GetValue("cert-file").String()
	keyFile := ctx.GetValue("key-file").String()
	if certFile != "" && keyFile != "" {
		ctx.Logger().Logf(logf.Info, "start listen tls %s", port)
		if err := http.ListenAndServeTLS(port, certFile, keyFile, handler{ctx: ctx}); err != nil {
			ctx.Logger().Logf(logf.Error, "listen: %s", err.Error())
			return false, nil
		}
		ctx.Logger().Logf(logf.Info, "listen complete")
		return false, nil
	}
	ctx.Logger().Logf(logf.Info, "start listen %s", port)
	if err := http.ListenAndServe(port, handler{ctx: ctx}); err != nil {
		ctx.Logger().Logf(logf.Error, "listen: %s", err.Error())
		return false, err
	}
	ctx.Logger().Logf(logf.Info, "listen complete")
	return false, nil
}

type handler struct {
	ctx *ngin.Context
}

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := h.ctx.Folk()
	withRequest(ctx, req)
	var ok bool
	var err error
	for _, stmt := range h.ctx.NextStmts() {
		if ok, err = stmt.Execute(ctx); err != nil || !ok {
			break
		}
	}
	keys := ctx.GetAttr("response.header.*").Slice()
	for _, key := range keys {
		w.Header().Set(key.String(), ctx.GetValue("response.header."+key.String()).String())
	}
	code := 200
	if c := ctx.GetValue("response.code").Int(); c != 0 {
		code = int(c)
	}
	w.WriteHeader(code)
	body := ctx.GetValue("response.body")
	if _, ok := body.(ngin.Null); !ok {
		w.Write(body.Bytes())
		return
	}
	if f := ctx.GetValuedFunc("responseBody"); f != nil {
		w.Write(f(ctx).Bytes())
	}
}
