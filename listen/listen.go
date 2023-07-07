// Copyright (c) 2023 Yang,Zhong
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package listen

import (
	"crypto/tls"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func Init(ctx *ngin.Context) {
	listener := listener{}
	ctx.BindFunc("listen", listener.listen)
	ctx.BindFunc("backend", listener.backend)
	ctx.BindFunc("call", listener.call)
	ctx.BindFunc("forward", listener.call)
}

type listener struct{}

func (listener) listen(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	network := "tcp"
	addr := ""
	protocol := "http"
	certFile := ctx.GetValue("cert-file").String()
	keyFile := ctx.GetValue("key-file").String()
	if len(args) < 1 {
		ctx.Logger().Logf(logf.Fatal, "you should provide an address after listen. example: listen :6000, listen tcp :6000, listen udp :6000")
		return false, nil
	} else if len(args) == 1 {
		addr = args[0].String()
	} else if len(args) <= 2 {
		network = args[0].String()
		addr = args[1].String()
	} else if len(args) <= 3 {
		network = args[0].String()
		addr = args[1].String()
		protocol = args[2].String()
	}
	listener, err := func() (net.Listener, error) {
		if idx := strings.Index(addr, ":"); idx < 0 {
			addr = ":" + addr
		}
		if certFile != "" && keyFile != "" {
			cer, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				ctx.Logger().Logf(logf.Fatal, "load key pair: %s", err.Error())
				return nil, err
			}
			config := &tls.Config{Certificates: []tls.Certificate{cer}}
			return tls.Listen(network, addr, config)
		}
		return net.Listen(network, addr)
	}()
	if err != nil {
		return false, err
	}
	ctx.Logger().Logf(logf.Info, "listen %s", addr)
	switch protocol {
	case "http":
		s := http.Server{Handler: httpHandler{ctx: ctx}}
		if err := s.Serve(listener); err != nil {
			ctx.Logger().Logf(logf.Error, "serve http: %s", err.Error())
		}
	case "ssh":
		// TODO implement
	}
	return false, nil
}

func (listener) backend(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	backends := []string{}
	for _, arg := range args {
		for _, v := range arg.Slice() {
			backends = append(backends, v.String())
		}
	}
	if len(backends) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide at least one backend")
		return false, nil
	}
	idx := rand.Intn(len(backends))
	u, err := url.Parse(backends[idx])
	if err != nil {
		ctx.Logger().Logf(logf.Error, "parse url: %s", err.Error())
		return true, err
	}
	ctx.Logger().Logf(logf.Info, "selected backend: %s", backends[idx])
	ctx.BindValue("host", ngin.String(u.Host))
	ctx.BindValue("scheme", ngin.String(u.Scheme))
	return true, nil
}

func (listener) RequestFromContext(ctx *ngin.Context) (*http.Request, error) {
	r := ctx.Get("request")
	if r == nil {
		ctx.Logger().Logf(logf.Error, "can't get request from context")
		return nil, errors.New("can't get request from context")
	}
	req := r.(*http.Request)
	req.RequestURI = ""
	req.URL.Scheme = ctx.GetValue("scheme").String()
	req.URL.Host = ctx.GetValue("host").String()
	keys := ctx.GetAttr("header").Slice()
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	for _, key := range keys {
		if k := key.String(); k != "" {
			v := ctx.GetValue("header." + k).String()
			req.Header.Set(k, v)
		}
	}
	keys = ctx.GetAttr("query").Slice()
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

func (h listener) call(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	req, err := h.RequestFromContext(ctx)
	if req.URL.Host == "" {
		ctx.Logger().Logf(logf.Error, "no backend found")
		return false, nil
	}
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
		ctx.BindValue("response.header."+k, ngin.String(resp.Header.Get(k)))
	}
	ctx.BindValue("response.code", ngin.Int(uint64(resp.StatusCode)))
	var body []byte
	ctx.BindValuedFunc("read-response-body", func(_ *ngin.Context, args ...ngin.Value) ngin.Value {
		if resp.Body == nil {
			return ngin.Bytes(body)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				ctx.Logger().Logf(logf.Error, "close body: %s", err.Error())
			}
			resp.Body = nil
		}()
		var err error
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			ctx.Logger().Logf(logf.Error, "read response body: %s", err.Error())
			return ngin.Null{}
		}
		return ngin.Bytes(body)
	})
	return true, nil
}

type httpHandler struct {
	ctx *ngin.Context
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := h.ctx.Folk()
	ctx.Declare("read-response-body")
	h.withRequest(ctx, req)
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
	if f := ctx.GetValuedFunc("read-response-body"); f != nil {
		w.Write(f(ctx).Bytes())
	}
}

func (h httpHandler) withRequest(ctx *ngin.Context, req *http.Request) *ngin.Context {
	ctx.Declare("path", "hash", "scheme", "host", "user-agent", "remote-addr", "method", "header", "response", "query")
	ctx.Put("request", req)
	ctx.BindValuedFunc("read-request-body", h.requestBody)
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

func (h httpHandler) requestBody(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
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
