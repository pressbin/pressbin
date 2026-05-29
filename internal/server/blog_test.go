package server_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pressbin.dev/pressbin/internal/server"
	"pressbin.dev/pressbin/internal/store"
	"pressbin.dev/pressbin/internal/testutil"
)

func TestBlog_publicPages(t *testing.T) {
	st := testutil.NewStore(t)
	syncKey := testutil.MustSyncKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	slug := "blog-pub-" + store.RandomKeySuffix(8)
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync",
		testutil.SyncJSON(slug, "Public Post", ""), syncKey)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("index status=%d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "text/html") && w.Header().Get("Content-Type") == "" {
		t.Error("expected HTML index")
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/p/"+slug, nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("post page status=%d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Public Post") {
		t.Error("post page missing title")
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/tags", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("tags status=%d", w.Code)
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/tag/sync", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("tag page status=%d", w.Code)
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/search?q=Public", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("search status=%d", w.Code)
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/feed.xml", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("rss status=%d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "<rss") {
		t.Error("expected RSS XML")
	}
}

func TestBlog_draftNotPublic(t *testing.T) {
	st := testutil.NewStore(t)
	syncKey := testutil.MustSyncKey(t, st)
	admin := testutil.MustAdminKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	slug := "blog-draft-" + store.RandomKeySuffix(8)
	body := []byte(`{"slug":"` + slug + `","content":"---\ntitle: Secret\ndate: 2026-05-01\nstatus: draft\n---\n\nHidden.\n"}`)
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", body, syncKey)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/p/"+slug, nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("draft public view status=%d", w.Code)
	}

	_ = testutil.DoRequest(t, h, http.MethodGet, "/api/admin/posts/"+slug, nil, admin)
}

func TestBlog_assets(t *testing.T) {
	h := testutil.NewServer(t, testutil.NewStore(t)).Handler()
	w := testutil.DoRequest(t, h, http.MethodGet, "/theme/style.css", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("assets status=%d", w.Code)
	}
}

func TestBlog_contentStaticImages(t *testing.T) {
	root := t.TempDir()
	imgDir := filepath.Join(root, "images")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "test.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := testutil.TestConfig()
	cfg.Assets.Path = root
	st := testutil.NewStore(t)
	h := server.New(st, cfg, testutil.TestAssets()).Handler()
	w := testutil.DoRequest(t, h, http.MethodGet, "/assets/images/test.svg", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("content image status=%d body=%s", w.Code, w.Body.String())
	}
}
