// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package listen

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func Init(ctx *ngin.Context) {
	ctx.BindFunc("listen", listen)
	ctx.BindFunc("backend", backend)
	ctx.BindFunc("call", call)
	ctx.BindFunc("forward", call)
}

func listen(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	listener := HttpListener{}
	return listener.Listen(ctx, args...)
}

func backend(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	listener := HttpListener{}
	return listener.Backend(ctx, args...)
}

func call(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	listener := HttpListener{}
	return listener.Call(ctx, args...)
}

type HttpListener struct{}

func (listener *HttpListener) withRequest(ctx *ngin.Context, req *http.Request) *ngin.Context {
	ctx.Declare("path", "hash", "scheme", "host", "user-agent", "remote-addr", "method", "header", "response", "query")
	ctx.Put("request", req)
	ctx.BindValuedFunc("body", listener.requestBody)
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

func (HttpListener) requestBody(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
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

func (HttpListener) Listen(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Fatal, "you shold provide a port after listen")
		return false, nil
	}
	port := ":" + args[0].String()
	certFile := ctx.GetValue("cert-file").String()
	keyFile := ctx.GetValue("key-file").String()
	if certFile != "" && keyFile != "" {
		ctx.Logger().Logf(logf.Info, "start listen tls %s", port)
		if err := http.ListenAndServeTLS(port, certFile, keyFile, httpHandler{ctx: ctx}); err != nil {
			ctx.Logger().Logf(logf.Error, "listen: %s", err.Error())
			return false, nil
		}
		ctx.Logger().Logf(logf.Info, "listen complete")
		return false, nil
	}
	ctx.Logger().Logf(logf.Info, "start listen %s", port)
	if err := http.ListenAndServe(port, httpHandler{ctx: ctx}); err != nil {
		ctx.Logger().Logf(logf.Error, "listen: %s", err.Error())
		return false, err
	}
	ctx.Logger().Logf(logf.Info, "listen complete")
	return false, nil
}

func (HttpListener) Backend(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
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

func (HttpListener) RequestFromContext(ctx *ngin.Context) (*http.Request, error) {
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

func (h HttpListener) Call(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	req, err := h.RequestFromContext(ctx)
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

type httpHandler struct {
	ctx *ngin.Context
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := h.ctx
	listener := HttpListener{}
	listener.withRequest(ctx, req)
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
