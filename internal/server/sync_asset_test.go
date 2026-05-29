package server_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/server"
	"pressbin.dev/pressbin/internal/testutil"
)

func TestSyncAsset_createUpdateDelete(t *testing.T) {
	root := t.TempDir()
	cfg := testutil.TestConfig()
	cfg.Assets.Path = root

	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	h := server.New(st, cfg, testutil.TestAssets(), "dev").Handler()

	body, _ := json.Marshal(map[string]string{
		"path":           "images/test.svg",
		"content_base64": base64.StdEncoding.EncodeToString([]byte("<svg/>")),
	})
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync/asset", body, key)
	if w.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["action"] != "created" {
		t.Errorf("action = %q", resp["action"])
	}

	got, err := os.ReadFile(filepath.Join(root, "images", "test.svg"))
	if err != nil || string(got) != "<svg/>" {
		t.Fatalf("file on disk: %v %q", err, got)
	}

	w = testutil.DoRequest(t, h, http.MethodPost, "/api/sync/asset", body, key)
	if w.Code != http.StatusOK {
		t.Fatalf("update status=%d", w.Code)
	}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["action"] != "updated" {
		t.Errorf("action = %q", resp["action"])
	}

	w = testutil.DoRequest(t, h, http.MethodDelete, "/api/sync/asset/images/test.svg", nil, key)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "images", "test.svg")); !os.IsNotExist(err) {
		t.Errorf("file still exists: %v", err)
	}
}

func TestSyncAsset_requiresStoragePath(t *testing.T) {
	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	body, _ := json.Marshal(map[string]string{
		"path":           "images/x.svg",
		"content_base64": base64.StdEncoding.EncodeToString([]byte("x")),
	})
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync/asset", body, key)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestSyncAsset_rejectsTraversal(t *testing.T) {
	root := t.TempDir()
	cfg := testutil.TestConfig()
	cfg.Assets.Path = root
	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	h := server.New(st, cfg, testutil.TestAssets(), "dev").Handler()

	body, _ := json.Marshal(map[string]string{
		"path":           "../escape.svg",
		"content_base64": base64.StdEncoding.EncodeToString([]byte("x")),
	})
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync/asset", body, key)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSyncAsset_allowsFontsPrefix(t *testing.T) {
	root := t.TempDir()
	cfg := testutil.TestConfig()
	cfg.Assets.Path = root

	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	h := server.New(st, cfg, testutil.TestAssets(), "dev").Handler()

	body, _ := json.Marshal(map[string]string{
		"path":           "fonts/custom.woff2",
		"content_base64": base64.StdEncoding.EncodeToString([]byte("x")),
	})
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync/asset", body, key)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "fonts", "custom.woff2")); err != nil {
		t.Fatalf("file on disk: %v", err)
	}
}
