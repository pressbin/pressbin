package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"pressbin.dev/pressbin/internal/store"
	"pressbin.dev/pressbin/internal/testutil"
)

func TestAdmin_postsCRUD(t *testing.T) {
	st := testutil.NewStore(t)
	admin := testutil.MustAdminKey(t, st)
	syncKey := testutil.MustSyncKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	slug := "admin-crud-" + store.RandomKeySuffix(8)
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", testutil.SyncJSON(slug, "Admin Post", ""), syncKey)
	if w.Code != http.StatusOK {
		t.Fatalf("sync setup: %s", w.Body.String())
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/api/admin/posts/"+slug, nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("get status=%d", w.Code)
	}

	w = testutil.DoRequest(t, h, http.MethodPatch, "/api/admin/posts/"+slug+"/status",
		[]byte(`{"status":"draft"}`), admin)
	if w.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%q", w.Code, w.Body.String())
	}

	w = testutil.DoRequest(t, h, http.MethodDelete, "/api/admin/posts/"+slug, nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status=%d", w.Code)
	}
}

func TestAdmin_updatePost(t *testing.T) {
	st := testutil.NewStore(t)
	syncKey := testutil.MustSyncKey(t, st)
	admin := testutil.MustAdminKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	slug := "admin-upd-" + store.RandomKeySuffix(8)
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", testutil.SyncJSON(slug, "Original", ""), syncKey)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}

	md := "---\ntitle: Revised Title\ndate: 2026-05-01\n---\n\n# Revised Title\n\nUpdated body.\n"
	update, _ := json.Marshal(map[string]string{
		"content": md,
		"title":   "Revised Title",
	})
	w = testutil.DoRequest(t, h, http.MethodPut, "/api/admin/posts/"+slug, update, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", w.Code, w.Body.String())
	}
	got, _ := st.GetPost(slug)
	if got.Title != "Revised Title" {
		t.Errorf("title = %q", got.Title)
	}
}

func TestAdmin_keysAndSettings(t *testing.T) {
	st := testutil.NewStore(t)
	admin := testutil.MustAdminKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	w := testutil.DoRequest(t, h, http.MethodPost, "/api/admin/keys",
		[]byte(`{"label":"ci-sync","type":"sync"}`), admin)
	if w.Code != http.StatusOK {
		t.Fatalf("create key status=%d", w.Code)
	}
	var keyResp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&keyResp); err != nil {
		t.Fatal(err)
	}
	if keyResp["key"] == "" {
		t.Fatal("expected raw key in response")
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/api/admin/keys", nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("list keys status=%d", w.Code)
	}

	w = testutil.DoRequest(t, h, http.MethodPut, "/api/admin/settings",
		[]byte(`{"site_title":"CI Blog","posts_per_page":3}`), admin)
	if w.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", w.Code, w.Body.String())
	}
	var settings map[string]any
	_ = json.NewDecoder(w.Body).Decode(&settings)
	if settings["site_title"] != "CI Blog" {
		t.Errorf("site_title = %v", settings["site_title"])
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/api/admin/stats", nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("stats status=%d", w.Code)
	}
}

func TestAdmin_createKey_invalidType(t *testing.T) {
	st := testutil.NewStore(t)
	admin := testutil.MustAdminKey(t, st)
	w := testutil.DoRequest(t, testutil.NewServer(t, st).Handler(), http.MethodPost, "/api/admin/keys",
		[]byte(`{"type":"robot"}`), admin)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}
