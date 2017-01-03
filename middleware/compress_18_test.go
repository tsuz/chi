// +build go1.8

package middleware

import (
	"crypto/tls"
	"io"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

// TestGenericWrapperHTTP2 tests that when an http2 request is wrapped using mkGenericWrapper, it still supports the
// various interfaces that a http.http2responseWriter implements.
func TestGenericWrapperHTTP2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, cn := w.(http.CloseNotifier)
		if !cn {
			t.Fatal("request should have been a http.CloseNotifier")
		}
		_, fl := w.(http.Flusher)
		if !fl {
			t.Fatal("request should have been a http.Flusher")
		}
		_, hj := w.(http.Hijacker)
		if hj {
			t.Fatal("request should not have been a http.Hijacker")
		}
		_, rf := w.(io.ReaderFrom)
		if rf {
			t.Fatal("request should not have been a io.ReaderFrom")
		}
		_, ps := w.(http.Pusher)
		if !ps {
			t.Fatal("request should have been a http.Pusher")
		}

		w.Write([]byte("OK"))
	})

	server := http.Server{
		Addr:    ":7072",
		Handler: identityMiddleware(handler),
	}
	// By serving over TLS, we get HTTP2 requests
	go server.ListenAndServeTLS("./testing/cert.pem", "./testing/key.pem")
	defer server.Close()
	// We need the server to start before making the request
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: &tls.Config{
				// The certificates we are using are self signed
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get("https://localhost:7072")
	if err != nil {
		t.Fatalf("could not get server: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("non 200 response: %v", resp.StatusCode)
	}
}
