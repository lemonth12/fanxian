package jd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSign(t *testing.T) {
	c := NewClient("key", "secret", "site", "pid")
	sig := c.sign(map[string]string{
		"method":  "test",
		"app_key": "key",
	})
	if sig == "" {
		t.Error("sign should not be empty")
	}
	// Same input should produce same sign
	sig2 := c.sign(map[string]string{
		"app_key": "key",
		"method":  "test",
	})
	if sig != sig2 {
		t.Error("sign should be deterministic regardless of map order")
	}
}

func TestSign_Empty(t *testing.T) {
	c := NewClient("key", "secret", "site", "pid")
	sig := c.sign(map[string]string{})
	if sig == "" {
		t.Error("sign with empty params should not be empty")
	}
}

func TestCall_Success(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				resp := APIResponse{
					Code:    "0",
					Message: "success",
					Result:  json.RawMessage(`{"data":"ok"}`),
				}
				body, _ := json.Marshal(resp)
				w.Header().Set("Content-Type", "application/json")
				w.Write(body)
			}))
	defer srv.Close()

	c := NewClient("key", "secret", "site", "pid")
	c.BaseURL = srv.URL

	resp, err := c.call("test.method", map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != "0" {
		t.Errorf("code = %s, want 0", resp.Code)
	}
}

func TestCall_NonRetryableError(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				resp := APIResponse{
					Code:    "2000",
					Message: "auth failed",
				}
				body, _ := json.Marshal(resp)
				w.Header().Set("Content-Type", "application/json")
				w.Write(body)
			}))
	defer srv.Close()

	c := NewClient("key", "secret", "site", "pid")
	c.BaseURL = srv.URL

	_, err := c.call("test.method", map[string]string{})
	if err == nil {
		t.Error("expected error for non-retryable code")
	}
}

func TestNewClient_HasDefaults(t *testing.T) {
	c := NewClient("key", "secret", "site", "pid")
	if c.BaseURL != baseURL {
		t.Errorf("BaseURL = %s, want %s", c.BaseURL, baseURL)
	}
	if c.HTTP == nil {
		t.Error("HTTP client should be set")
	}
}
