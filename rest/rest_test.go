package rest_test

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"gozerosource/rest"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ServiceContext struct {
	Config rest.RestConf
}

func NewServiceContext(c rest.RestConf) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func IndexHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(">>>>>>>>>>>>>>>>")
		w.Write([]byte("pong"))
	}
}

func ServerStart() {
	c := rest.RestConf{
		Host:         "127.0.0.1",
		Port:         8081,
		MaxConns:     100,
		MaxBytes:     1048576,
		Timeout:      1000,
		CpuThreshold: 800,
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{
				Path: "./",
			},
		},
	}
	server := rest.MustNewServer(c)
	// defer server.Stop()
	ctx := NewServiceContext(c)
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ping",
		Handler: IndexHandler(ctx),
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func Test_Rest(t *testing.T) {
	go func() {
		ServerStart()
	}()
	time.Sleep(1 * time.Second)

	for _, tt := range [...]struct {
		name string

		method, uri string
		body        io.Reader

		want     *http.Request
		wantBody string
	}{
		{
			name:   "GET with full URL",
			method: "GET",
			uri:    "http://127.0.0.1:8081/ping",
			body:   nil,
			want: &http.Request{
				Method: "GET",
				Host:   "127.0.0.1:8081",
				URL: &url.URL{
					Scheme:  "http",
					Path:    "/ping",
					RawPath: "/ping",
					Host:    "127.0.0.1:8081",
				},
				Header:     http.Header{},
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				RemoteAddr: "192.0.2.1:1234",
				RequestURI: "http://127.0.0.1:8081",
			},
			wantBody: "pong",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRequest(tt.method, tt.uri, tt.body)
			slurp, err := io.ReadAll(got.Body)
			if err != nil {
				t.Errorf("ReadAll: %v", err)
			}
			fmt.Printf(" --> %s", string(slurp))
			if string(slurp) != tt.wantBody {
				t.Errorf("Body = %q; want %q", slurp, tt.wantBody)
			}
		})
	}
}

func NewRequest(method, target string, body io.Reader) *http.Request {
	if method == "" {
		method = "GET"
	}
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(method + " " + target + " HTTP/1.0\r\n\r\n")))
	if err != nil {
		panic("invalid NewRequest arguments; " + err.Error())
	}

	// HTTP/1.0 was used above to avoid needing a Host field. Change it to 1.1 here.
	req.Proto = "HTTP/1.1"
	req.ProtoMinor = 1
	req.Close = false

	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
		default:
			req.ContentLength = -1
		}
		if rc, ok := body.(io.ReadCloser); ok {
			req.Body = rc
		} else {
			req.Body = io.NopCloser(body)
		}
	}

	// 192.0.2.0/24 is "TEST-NET" in RFC 5737 for use solely in
	// documentation and example source code and should not be
	// used publicly.
	req.RemoteAddr = "192.0.2.1:1234"

	if req.Host == "" {
		req.Host = "example.com"
	}

	if strings.HasPrefix(target, "https://") {
		req.TLS = &tls.ConnectionState{
			Version:           tls.VersionTLS12,
			HandshakeComplete: true,
			ServerName:        req.Host,
		}
	}

	return req
}

func TestHandlePost(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/topic/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	reader := strings.NewReader(`{"title":"The Go Standard Library","content":"It contains many packages."}`)
	r, _ := http.NewRequest(http.MethodPost, "/topic/", reader)

	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response code is %v", resp.StatusCode)
	}
}
