package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// func Serve(cfg *Config) error {
// 	fmt.Printf("%+v\n", cfg)
// 	srv := newServer(cfg)
// 	return srv.start()
// }

type proxy struct {
	cfg      *Config
	ln       net.Listener
	rp       *httputil.ReverseProxy
	requests chan struct{}
	unpause  chan struct{}
	errStr   string
}

func newProxy(cfg *Config) *proxy {
	url, err := url.Parse(fmt.Sprintf("%s:%v", "http://localhost", cfg.AppPort))
	if err != nil {
		log.Fatal(err)
	}

	rp := httputil.NewSingleHostReverseProxy(url)
	rp.ErrorLog = log.New(ioutil.Discard, "", 0)

	p := &proxy{
		cfg:      cfg,
		rp:       rp,
		requests: make(chan struct{}),
		unpause:  make(chan struct{}),
	}
	return p
}

func (p *proxy) start(ready chan error) error {
	srv := &http.Server{
		Handler: p,
		BaseContext: func(ln net.Listener) context.Context {
			return context.Background()
		},
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "localhost", p.cfg.ProxyPort))
	if err != nil {
		return err
	}
	defer ln.Close()
	p.ln = ln
	if ready != nil {
		ready <- nil
	}

	p.cfg.Printf("proxying requests on %s to :%d", ln.Addr(), p.cfg.AppPort)
	return srv.Serve(ln)
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.requests <- struct{}{}
	<-p.unpause

	ctx, cancel := context.WithTimeout(r.Context(), p.cfg.Timeout)
	defer cancel()

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Failed to read request body\n"))
		ignoreError(err)
		return
	}

	s := string(b)

	for {
		if ok := p.forward(w, r, s); ok {
			return
		}

		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			p.cfg.Print("timeout reached")
			w.WriteHeader(http.StatusBadGateway)
			_, err := w.Write([]byte("Connection Refused\n"))
			ignoreError(err)
			return
		}
	}
}

func (p *proxy) forward(w http.ResponseWriter, r *http.Request, body string) bool {
	if len(p.errStr) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(p.errStr))
		ignoreError(err)
		return true
	}

	r.Body = &stringReader{Reader: strings.NewReader(body)}

	writer := &proxyWriter{res: w}
	p.rp.ServeHTTP(writer, r)
	// fmt.Println("proxyWriter.status", writer.status)

	// If the request is "successful" - as in the server responded in
	// some way, return the response to the client.
	return writer.status != http.StatusBadGateway
}

func (p *proxy) setError(err error) {
	p.cfg.Debug("proxy: error mode")
	p.errStr = err.Error()
}

func (p *proxy) clearError() {
	p.errStr = ""
}

// Wrapper around http.ResponseWriter. Since the proxy works rather naively -
// it just retries requests over and over until it gets a response from the app
// server - we can't use the ResponseWriter that is passed to the handler
// because you cannot call WriteHeader multiple times.
type proxyWriter struct {
	res    http.ResponseWriter
	status int
}

func (w *proxyWriter) WriteHeader(status int) {
	// fmt.Println("WriteHeader", status)
	w.status = status
	if status == 502 {
		return
	}

	w.res.WriteHeader(status)
}

func (w *proxyWriter) Write(body []byte) (int, error) {
	return w.res.Write(body)
}

func (w *proxyWriter) Header() http.Header {
	return w.res.Header()
}

type stringReader struct {
	*strings.Reader
}

func (r *stringReader) Close() error { return nil }
