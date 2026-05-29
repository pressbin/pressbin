package syncclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/syncclient"
)

func TestPushPost_and_DeletePost(t *testing.T) {
	var gotSlug string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/sync":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotSlug, _ = body["slug"].(string)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/sync/hello":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	post := filepath.Join(dir, "posts", "hello.md")
	if err := os.MkdirAll(filepath.Dir(post), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ntitle: Hello\nslug: hello\n---\n\nBody\n"
	if err := os.WriteFile(post, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PRESSBIN_URL", srv.URL)
	t.Setenv("PRESSBIN_KEY", "pb_sync_test")

	c := &syncclient.Client{BaseURL: srv.URL, Key: "pb_sync_test", HTTP: srv.Client()}
	if err := c.PushPost(post); err != nil {
		t.Fatalf("PushPost: %v", err)
	}
	if gotSlug != "hello" {
		t.Fatalf("slug = %q", gotSlug)
	}
	if err := c.DeletePost("hello"); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
}

func TestPushAsset(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/sync/asset" {
			var body struct {
				Path string `json:"path"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotPath = body.Path
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	asset := filepath.Join(dir, "assets", "images", "a.svg")
	if err := os.MkdirAll(filepath.Dir(asset), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(asset, []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := &syncclient.Client{BaseURL: srv.URL, Key: "k", HTTP: srv.Client()}
	if err := c.PushAsset(asset); err != nil {
		t.Fatalf("PushAsset: %v", err)
	}
	if gotPath != "images/a.svg" {
		t.Fatalf("path = %q", gotPath)
	}
}
