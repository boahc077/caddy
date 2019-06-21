package markdown

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"

	"gopkg.in/russross/blackfriday.v2"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(caddy.Module{
		Name: "http.middleware.markdown",
		New:  func() interface{} { return new(Markdown) },
	})
}

// Markdown is a middleware for rendering a Markdown response body.
type Markdown struct {
}

func (m Markdown) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	rr := caddyhttp.NewResponseRecorder(w, buf)

	err := next.ServeHTTP(rr, r)
	if err != nil {
		return err
	}

	output := blackfriday.Run(buf.Bytes())

	w.Header().Set("Content-Length", strconv.Itoa(len(output)))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Del("Accept-Ranges") // we don't know ranges for dynamically-created content
	w.Header().Del("Etag")          // don't know a way to quickly generate etag for dynamic content
	w.Header().Del("Last-Modified") // useless for dynamic content since it's always changing

	w.WriteHeader(rr.Status())
	w.Write(output)

	return nil
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Interface guard
var _ caddyhttp.MiddlewareHandler = (*Markdown)(nil)
