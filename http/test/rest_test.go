package rest_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_Rest(t *testing.T) {
	for _, tt := range [...]struct {
		name string

		method, uri string
		body        io.Reader

		want     *http.Request
		wantBody string
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
			body, err := NewRequest(tt.method, tt.uri, tt.body)
			if err != nil {
				t.Errorf("ReadAll: %v", err)
			}
			if string(body) != tt.wantBody {
				t.Errorf("Body = %q; want %q", body, tt.wantBody)
			}
		})
	}
}

func NewRequest(method, url string, bodyRow io.Reader) ([]byte, error) {
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
