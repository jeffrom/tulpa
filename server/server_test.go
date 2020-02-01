package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestServerGet(t *testing.T) {
	mockCommand()
	defer resetCommand()
	cfg := newTestConfig()
	srv := newTestAppServer(cfg, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	defer srv.Close()

	s, errC := newTestServer(cfg, "cool")
	defer checkNoServerError(t, errC)

	res, err := http.Get(fmt.Sprintf("http://%s", s.Addr()))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Fatal("expected 200, got", res.StatusCode)
	}
	// b, _ := httputil.DumpResponse(res, true)
	// fmt.Println(string(b))
}

func TestServerPost(t *testing.T) {
	mockCommand()
	defer resetCommand()
	cfg := newTestConfig()

	var n int32 = 0
	srv := newTestAppServer(cfg, func(w http.ResponseWriter, r *http.Request) {
		// simulate gateway timeout for the first request. this tests we're
		// managing the request.Body correctly.
		if atomic.CompareAndSwapInt32(&n, 0, 1) {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(200)
	})
	defer srv.Close()

	s, errC := newTestServer(cfg, "cool")
	defer checkNoServerError(t, errC)

	res, err := http.Post(fmt.Sprintf("http://%s", s.Addr()), "text/plain", strings.NewReader("cool post"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Fatal("expected 200, got", res.StatusCode)
	}
}

func checkNoServerError(t testing.TB, errC chan error) {
	t.Helper()

	select {
	case err := <-errC:
		t.Fatal(err)
	default:
	}
}

func newTestConfig() *Config {
	return &Config{
		AppPort: 0,
		Timeout: 1 * time.Second,
		// TODO can test output this way
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func newTestServer(cfg *Config, args ...string) (*Server, chan error) {
	s := New(cfg, args)
	errC := s.GoStart()

	hostport := s.Addr().String()
	port, err := strconv.Atoi(strings.SplitN(hostport, ":", 2)[1])
	if err != nil {
		panic(err)
	}
	cfg.ProxyPort = port
	return s, errC
}

func newTestAppServer(cfg *Config, fn http.HandlerFunc) *httptest.Server {
	srv := httptest.NewServer(fn)

	hostport := srv.Listener.Addr().String()
	port, err := strconv.Atoi(strings.SplitN(hostport, ":", 2)[1])
	if err != nil {
		panic(err)
	}
	cfg.AppPort = port
	return srv
}
