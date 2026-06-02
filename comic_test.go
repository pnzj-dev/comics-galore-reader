package cgreader

import (
	"errors"
	"testing"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s       string
		substrs []string
		want    bool
	}{
		{"encrypted archive", []string{"encrypted", "password"}, true},
		{"wrong password", []string{"encrypted", "password"}, true},
		{"all good", []string{"encrypted", "password"}, false},
		{"CRC mismatch", []string{"CRC"}, true},
		{"", []string{"a", "b"}, false},
		{"abc", nil, false},
	}
	for _, tt := range tests {
		got := containsAny(tt.s, tt.substrs...)
		if got != tt.want {
			t.Errorf("containsAny(%q, %v) = %v, want %v", tt.s, tt.substrs, got, tt.want)
		}
	}
}

func TestIsEncryptionError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{errors.New("encrypted"), true},
		{errors.New("wrong password"), true},
		{errors.New("CRC error"), true},
		{errors.New("checksum mismatch"), true},
		{errors.New("some other error"), false},
		{nil, false},
	}
	for _, tt := range tests {
		got := isEncryptionError(tt.err)
		if got != tt.want {
			t.Errorf("isEncryptionError(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestNewComicReaderDefaults(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.currentPage != 0 {
		t.Errorf("expected currentPage=0, got %d", r.currentPage)
	}
	if r.passwordURL != "" {
		t.Errorf("expected empty passwordURL, got %q", r.passwordURL)
	}
	if !r.cacheEnabled {
		t.Error("expected cacheEnabled=true by default")
	}
}

func TestNewComicReaderWithOptions(t *testing.T) {
	r := New(
		WithPasswordURL("/custom/password"),
		WithCacheEnabled(false),
	)
	if r.passwordURL != "/custom/password" {
		t.Errorf("expected passwordURL=/custom/password, got %q", r.passwordURL)
	}
	if r.cacheEnabled {
		t.Error("expected cacheEnabled=false")
	}
}

func TestComicReaderNoArchive(t *testing.T) {
	r := New()
	if r.PageCount() != 0 {
		t.Errorf("expected PageCount=0, got %d", r.PageCount())
	}
	if r.CurrentPage() != 0 {
		t.Errorf("expected CurrentPage=0, got %d", r.CurrentPage())
	}
	if r.Title() != "" {
		t.Errorf("expected empty Title, got %q", r.Title())
	}
	if r.Filename() != "" {
		t.Errorf("expected empty Filename, got %q", r.Filename())
	}
	if r.Next() {
		t.Error("expected Next()=false with no archive")
	}
	if r.Prev() {
		t.Error("expected Prev()=false with no archive")
	}
	if _, err := r.GetPage(0); err == nil {
		t.Error("expected error from GetPage() with no archive")
	}
	if err := r.SetCurrentPage(0); err == nil {
		t.Error("expected error from SetCurrentPage() with no archive")
	}
}

func TestSetPasswordURL(t *testing.T) {
	r := New()
	r.SetPasswordURL("/new/password")
	if r.passwordURL != "/new/password" {
		t.Errorf("expected passwordURL=/new/password, got %q", r.passwordURL)
	}
}
