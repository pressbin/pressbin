package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"pressbin.dev/pressbin/internal/store"
	"pressbin.dev/pressbin/internal/testutil"
)

func syncBody(slug, title string) []byte {
	content := fmt.Sprintf(`---
title: %s
date: 2026-05-10
tags: [sync, test]
---

# %s

Synced body with **markdown**.
`, title, title)
	return testutil.SyncJSON(slug, title, content)
}

func TestSync_createAndUpdate(t *testing.T) {
	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	srv := testutil.NewServer(t, st)
	h := srv.Handler()

	slug := "sync-test-" + store.RandomKeySuffix(8)
	body := syncBody(slug, "Sync Title")

	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", body, key)
	if w.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["action"] != "created" {
		t.Errorf("action = %q", resp["action"])
	}

	w = testutil.DoRequest(t, h, http.MethodPost, "/api/sync", body, key)
	if w.Code != http.StatusOK {
		t.Fatalf("update status=%d", w.Code)
	}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["action"] != "updated" {
		t.Errorf("action = %q", resp["action"])
	}

	got, err := st.GetPost(slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Sync Title" {
		t.Errorf("title = %q", got.Title)
	}
	if len(got.Tags) < 2 {
		t.Errorf("tags = %v", got.Tags)
	}
}

func TestSync_validationErrors(t *testing.T) {
	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", []byte(`{}`), key)
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty body status = %d", w.Code)
	}

}

func TestSync_requiresAuth(t *testing.T) {
	st := testutil.NewStore(t)
	h := testutil.NewServer(t, st).Handler()
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", syncBody("no-auth", "T"), "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSyncDelete(t *testing.T) {
	st := testutil.NewStore(t)
	key := testutil.MustSyncKey(t, st)
	h := testutil.NewServer(t, st).Handler()

	slug := "sync-del-" + store.RandomKeySuffix(8)
	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", syncBody(slug, "Del"), key)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}

	path := "/api/sync/" + slug
	w = testutil.DoRequest(t, h, http.MethodDelete, path, nil, key)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", w.Code, w.Body.String())
	}
	ok, _ := st.PostExists(slug)
	if ok {
		t.Error("post still exists")
	}

	w = testutil.DoRequest(t, h, http.MethodDelete, path, nil, key)
	if w.Code != http.StatusNotFound {
		t.Fatalf("second delete status=%d", w.Code)
	}
}
