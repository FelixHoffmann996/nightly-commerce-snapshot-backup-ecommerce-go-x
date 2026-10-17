package snapshot

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestPresignPutRetriesRateLimitAndDecodesEnvelope(t *testing.T) {
	attempts := 0
	client := NewStorageClient("test-key")
	client.BaseURL = "https://example.test"
	client.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if req.Method != http.MethodPost {
			t.Fatalf("method = %s", req.Method)
		}
		if attempts == 1 {
			return response(429, `{"ok":false,"error":{"message":"retry later"}}`, "2"), nil
		}
		return response(200, `{"ok":true,"data":{"url":"https://storage.test/put"},"metadata":{}}`, ""), nil
	})}
	var slept time.Duration
	client.Sleep = func(_ context.Context, delay time.Duration) error { slept = delay; return nil }

	got, err := client.PresignPut(context.Background(), "commerce", "2026-09-06.json", "snapshot-2026-09-06", 512)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://storage.test/put" || attempts != 2 || slept != 2*time.Second {
		t.Fatalf("url=%q attempts=%d slept=%s", got, attempts, slept)
	}
}

func response(status int, body, retryAfter string) *http.Response {
	header := make(http.Header)
	if retryAfter != "" {
		header.Set("Retry-After", retryAfter)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body))}
}
