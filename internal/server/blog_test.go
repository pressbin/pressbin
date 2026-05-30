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
	for _, path := range []string{"/theme/style.css", "/theme/htmx.min.js", "/theme/fonts/dm-sans.woff2"} {
		w := testutil.DoRequest(t, h, http.MethodGet, path, nil, "")
		if w.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, w.Code)
		}
		if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age=31536000") {
			t.Errorf("%s Cache-Control=%q, want long-lived cache", path, cc)
		}
	}
}

func TestBlog_htmxOnlyOnIndex(t *testing.T) {
	st := testutil.NewStore(t)
	syncKey := testutil.MustSyncKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	slug := "htmx-only-" + store.RandomKeySuffix(8)
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync",
		testutil.SyncJSON(slug, "HTMX Test", ""), syncKey)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/", nil, "")
	body := w.Body.String()
	if !strings.Contains(body, "<style>") || !strings.Contains(body, "Georgia, serif") {
		t.Error("index missing inlined critical CSS")
	}
	if !strings.Contains(body, `rel="preload"`) || !strings.Contains(body, "dm-sans.woff2") {
		t.Error("index missing font preload")
	}
	if !strings.Contains(body, `/theme/htmx.min.js`) {
		t.Error("index missing self-hosted htmx")
	}
	if !strings.Contains(body, `defer`) {
		t.Error("index htmx should be deferred")
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/p/"+slug, nil, "")
	body = w.Body.String()
	if strings.Contains(body, "htmx") {
		t.Error("post page should not load htmx")
	}
	if strings.Contains(body, "fonts.googleapis.com") || strings.Contains(body, "unpkg.com") {
		t.Error("post page should not load third-party fonts or scripts")
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
	h := server.New(st, cfg, testutil.TestAssets(), "dev").Handler()
	w := testutil.DoRequest(t, h, http.MethodGet, "/assets/images/test.svg", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("content image status=%d body=%s", w.Code, w.Body.String())
	}
}
