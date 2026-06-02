package cgreader

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchPassword(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mypassword"))
	}))
	defer ts.Close()

	pw, err := FetchPassword(ts.URL)
	if err != nil {
		t.Fatalf("FetchPassword failed: %v", err)
	}
	if pw != "mypassword" {
		t.Errorf("expected 'mypassword', got %q", pw)
	}
}

func TestFetchPasswordEmptyURL(t *testing.T) {
	_, err := FetchPassword("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestFetchPasswordNonOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	_, err := FetchPassword(ts.URL)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestFetchPasswordEmptyBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("  \n  "))
	}))
	defer ts.Close()

	_, err := FetchPassword(ts.URL)
	if err == nil {
		t.Fatal("expected error for empty password body")
	}
}

func TestFetchPasswordTrimsWhitespace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("  secretpwd\n"))
	}))
	defer ts.Close()

	pw, err := FetchPassword(ts.URL)
	if err != nil {
		t.Fatalf("FetchPassword failed: %v", err)
	}
	if pw != "secretpwd" {
		t.Errorf("expected 'secretpwd', got %q", pw)
	}
}
