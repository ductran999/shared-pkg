package httpclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	httpclient "github.com/ductran999/shared-pkg/transport/httpclient/v2"
)

func ExampleClient_Get() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":"hello"}`))
	}))
	defer srv.Close()

	client := httpclient.New()
	resp, err := client.Get(context.Background(), srv.URL)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	var out struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return
	}
	fmt.Println(out.Message)
	// Output: hello
}

func ExampleClient_Post() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	payload, _ := json.Marshal(map[string]string{"name": "a"})

	client := httpclient.New()
	resp, err := client.Post(context.Background(), srv.URL,
		bytes.NewReader(payload),
		httpclient.WithContentType("application/json"),
	)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println(resp.StatusCode)
	// Output: 201
}

func ExampleWithRequestID() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.Header.Get("X-Request-ID")))
	}))
	defer srv.Close()

	client := httpclient.New()
	ctx := httpclient.WithRequestID(context.Background(), "req-123")

	resp, err := client.Get(ctx, srv.URL)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output: req-123
}
