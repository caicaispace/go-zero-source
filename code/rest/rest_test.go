package rest_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"gozerosource/code/rest"
)

func Test_Rest(t *testing.T) {
	go rest.ServerStart()
	time.Sleep(time.Millisecond)
	for _, tt := range [...]struct {
		name, method, uri string
		body              io.Reader
		want              *http.Request
		wantBody          string
	}{
		{
			name:   "GET with ping url",
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
				Header: http.Header{},
				Proto:  "HTTP/1.1",
			},
			wantBody: "pong",
		},
		{
			name:   "GET with check url",
			method: "GET",
			uri:    "http://127.0.0.1:8081/check",
			body:   nil,
			want: &http.Request{
				Method: "GET",
				Host:   "127.0.0.1:8081",
				URL: &url.URL{
					Scheme:  "http",
					Path:    "/check",
					RawPath: "/check",
					Host:    "127.0.0.1:8081",
				},
				Header: http.Header{},
				Proto:  "HTTP/1.1",
			},
			wantBody: "ok",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body, err := httpRequest(tt.method, tt.uri, tt.body)
			if err != nil {
				t.Errorf("ReadAll: %v", err)
			}
			if string(body) != tt.wantBody {
				t.Errorf("Body = %q; want %q", body, tt.wantBody)
			}
		})
	}
}

func httpRequest(method, url string, bodyRow io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, bodyRow)
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	rsp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
