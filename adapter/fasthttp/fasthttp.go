// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package fasthttp

import (
	"bytes"
	"errors"
	"github.com/GoAdminGroup/go-admin/adapter"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Fasthttp structure value is a Fasthttp GoAdmin adapter.
type Fasthttp struct {
	adapter.BaseAdapter
	ctx *fasthttp.RequestCtx
	app *fasthttprouter.Router
}

func init() {
	engine.Register(new(Fasthttp))
}

func (fast *Fasthttp) User(ci interface{}) (models.UserModel, bool) {
	return fast.GetUser(ci, fast)
}

func (fast *Fasthttp) Use(router interface{}, plugs []plugins.Plugin) error {
	return fast.GetUse(router, plugs, fast)
}

func (fast *Fasthttp) Content(ctx interface{}, getPanelFn types.GetPanelFn) {
	fast.GetContent(ctx, getPanelFn, fast)
}

type HandlerFunc func(ctx *fasthttp.RequestCtx) (types.Panel, error)

func Content(handler HandlerFunc) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		engine.Content(ctx, func(ctx interface{}) (types.Panel, error) {
			return handler(ctx.(*fasthttp.RequestCtx))
		})
	}
}

func (fast *Fasthttp) SetApp(app interface{}) error {
	var (
		eng *fasthttprouter.Router
		ok  bool
	)
	if eng, ok = app.(*fasthttprouter.Router); !ok {
		return errors.New("wrong parameter")
	}

	fast.app = eng
	return nil
}

func (fast *Fasthttp) AddHandler(method, path string, handlers context.Handlers) {
	fast.app.Handle(strings.ToUpper(method), path, func(c *fasthttp.RequestCtx) {
		httpreq := convertCtx(c)
		ctx := context.NewContext(httpreq)

		var params = make(map[string]string)
		c.VisitUserValues(func(i []byte, i2 interface{}) {
			if value, ok := i2.(string); ok {
				params[string(i)] = value
			}
		})

		for key, value := range params {
			if httpreq.URL.RawQuery == "" {
				httpreq.URL.RawQuery += strings.Replace(key, ":", "", -1) + "=" + value
			} else {
				httpreq.URL.RawQuery += "&" + strings.Replace(key, ":", "", -1) + "=" + value
			}
		}

		ctx.SetHandlers(handlers).Next()
		for key, head := range ctx.Response.Header {
			c.Response.Header.Set(key, head[0])
		}
		if ctx.Response.Body != nil {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(ctx.Response.Body)
			_, _ = c.WriteString(buf.String())
		}
		c.Response.SetStatusCode(ctx.Response.StatusCode)
	})
}

func convertCtx(ctx *fasthttp.RequestCtx) *http.Request {
	var r http.Request

	body := ctx.PostBody()
	r.Method = string(ctx.Method())
	r.Proto = "HTTP/1.1"
	r.ProtoMajor = 1
	r.ProtoMinor = 1
	r.RequestURI = string(ctx.RequestURI())
	r.ContentLength = int64(len(body))
	r.Host = string(ctx.Host())
	r.RemoteAddr = ctx.RemoteAddr().String()

	hdr := make(http.Header)
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			r.TransferEncoding = append(r.TransferEncoding, sv)
		default:
			hdr.Set(sk, sv)
		}
	})
	r.Header = hdr
	r.Body = &netHTTPBody{body}
	rURL, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		ctx.Logger().Printf("cannot parse requestURI %q: %s", r.RequestURI, err)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return &r
	}
	r.URL = rURL
	return &r
}

type netHTTPBody struct {
	b []byte
}

func (r *netHTTPBody) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func (r *netHTTPBody) Close() error {
	r.b = r.b[:0]
	return nil
}

func (fast *Fasthttp) Name() string {
	return "fasthttp"
}

func (fast *Fasthttp) SetContext(contextInterface interface{}) adapter.WebFrameWork {
	var (
		ctx *fasthttp.RequestCtx
		ok  bool
	)
	if ctx, ok = contextInterface.(*fasthttp.RequestCtx); !ok {
		panic("wrong parameter")
	}
	return &Fasthttp{ctx: ctx}
}

func (fast *Fasthttp) Redirect() {
	fast.ctx.Redirect(config.Get().Url("/login"), http.StatusFound)
}

func (fast *Fasthttp) SetContentType() {
	fast.ctx.Response.Header.Set("Content-Type", fast.HTMLContentType())
}

func (fast *Fasthttp) Write(body []byte) {
	_, _ = fast.ctx.Write(body)
}

func (fast *Fasthttp) GetCookie() (string, error) {
	return string(fast.ctx.Request.Header.Cookie(fast.CookieKey())), nil
}

func (fast *Fasthttp) Path() string {
	return string(fast.ctx.Path())
}

func (fast *Fasthttp) Method() string {
	return string(fast.ctx.Method())
}

func (fast *Fasthttp) FormParam() url.Values {
	f, _ := fast.ctx.MultipartForm()
	if f != nil {
		return f.Value
	}
	return url.Values{}
}

func (fast *Fasthttp) PjaxHeader() string {
	return string(fast.ctx.Request.Header.Peek(constant.PjaxHeader))
}
