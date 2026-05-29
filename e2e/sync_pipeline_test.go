package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pressbin.dev/pressbin/internal/store"
	"pressbin.dev/pressbin/internal/syncclient"
	"pressbin.dev/pressbin/internal/testutil"
)

func TestE2E_versionDiscovery_matchesServer(t *testing.T) {
	h := NewHarnessWithVersion(t, "v2.3.4")

	got, err := syncclient.FetchVersion(h.BaseURL(), h.Server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if got != "v2.3.4" {
		t.Fatalf("version = %q", got)
	}
}

func TestE2E_syncClientFromEnv_firstPushToPublicBlog(t *testing.T) {
	h := NewHarness(t)
	h.WithPressbinEnv()

	repo := InitGitRepo(t)
	slug := "e2e-hello-" + store.RandomKeySuffix(6)
	WriteRepoFile(t, repo, "posts/"+slug+".md", `---
title: E2E Hello
date: 2026-05-01
tags: [e2e, pressbin]
slug: `+slug+`
status: published
---

# Hello from git

Body for e2e.
`)
	WriteRepoFile(t, repo, "assets/images/logo.svg", `<svg xmlns="http://www.w3.org/2000/svg"/>`)
	GitCommitAll(t, repo, "initial content")

	client, err := syncclient.NewFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := syncclient.Run(client, repo); err != nil {
		t.Fatalf("Run: %v", err)
	}

	handler := h.Handler()
	w := testutil.DoRequest(t, handler, http.MethodGet, "/", nil, "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "E2E Hello") {
		t.Fatalf("index: status=%d missing title", w.Code)
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/p/"+slug, nil, "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Hello from git") {
		t.Fatalf("post page: status=%d", w.Code)
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/tag/e2e", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("tag page: status=%d", w.Code)
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/assets/images/logo.svg", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("asset HTTP: status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Fatal("asset body missing svg")
	}

	onDisk := filepath.Join(h.AssetsDir, "images", "logo.svg")
	if _, err := os.Stat(onDisk); err != nil {
		t.Fatalf("asset file on disk: %v", err)
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/api/admin/stats", nil, h.AdminKey)
	if w.Code != http.StatusOK {
		t.Fatalf("admin stats: status=%d", w.Code)
	}
	var stats map[string]any
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}
	if stats["last_sync_at"] == nil {
		t.Error("expected last_sync_at after sync")
	}
}

func TestE2E_incrementalGitPush_updatesAndDeletes(t *testing.T) {
	h := NewHarness(t)
	client := h.SyncClient()

	repo := InitGitRepo(t)
	slugKeep := "e2e-keep-" + store.RandomKeySuffix(6)
	slugGone := "e2e-gone-" + store.RandomKeySuffix(6)

	WriteRepoFile(t, repo, "posts/"+slugKeep+".md", "---\ntitle: Keep Me\nslug: "+slugKeep+"\ndate: 2026-05-01\nstatus: published\n---\n\nv1\n")
	WriteRepoFile(t, repo, "posts/"+slugGone+".md", "---\ntitle: Gone\nslug: "+slugGone+"\ndate: 2026-05-01\nstatus: published\n---\n\nbye\n")
	WriteRepoFile(t, repo, "assets/images/old.svg", "<svg id=\"old\"/>")
	GitCommitAll(t, repo, "v1")

	if err := syncclient.Run(client, repo); err != nil {
		t.Fatal(err)
	}

	WriteRepoFile(t, repo, "posts/"+slugKeep+".md", "---\ntitle: Keep Me Updated\nslug: "+slugKeep+"\ndate: 2026-05-01\nstatus: published\n---\n\nv2\n")
	_ = os.Remove(filepath.Join(repo, "posts", slugGone+".md"))
	_ = os.Remove(filepath.Join(repo, "assets/images/old.svg"))
	WriteRepoFile(t, repo, "assets/images/new.svg", "<svg id=\"new\"/>")
	GitCommitAll(t, repo, "v2")

	if err := syncclient.Run(client, repo); err != nil {
		t.Fatal(err)
	}

	handler := h.Handler()

	w := testutil.DoRequest(t, handler, http.MethodGet, "/p/"+slugKeep, nil, "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Keep Me Updated") {
		t.Fatalf("updated post: status=%d body=%q", w.Code, w.Body.String())
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/p/"+slugGone, nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("deleted post: status=%d want 404", w.Code)
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/assets/images/new.svg", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("new asset: status=%d", w.Code)
	}

	w = testutil.DoRequest(t, handler, http.MethodGet, "/assets/images/old.svg", nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("deleted asset: status=%d want 404", w.Code)
	}

	exists, _ := h.Store.PostExists(slugGone)
	if exists {
		t.Error("deleted slug still in database")
	}
}

func TestE2E_draftPost_syncClientHiddenFromPublic(t *testing.T) {
	h := NewHarness(t)
	repo := InitGitRepo(t)
	slug := "e2e-draft-" + store.RandomKeySuffix(6)

	WriteRepoFile(t, repo, "posts/"+slug+".md", `---
title: Secret Draft
slug: `+slug+`
date: 2026-05-01
status: draft
---

Hidden.
`)
	GitCommitAll(t, repo, "draft")

	if err := syncclient.Run(h.SyncClient(), repo); err != nil {
		t.Fatal(err)
	}

	w := testutil.DoRequest(t, h.Handler(), http.MethodGet, "/p/"+slug, nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("draft public page: status=%d want 404", w.Code)
	}

	got, err := h.Store.GetPost(slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "draft" {
		t.Fatalf("status = %q", got.Status)
	}
}

func TestE2E_authBoundary_syncKeyCannotReachAdmin(t *testing.T) {
	h := NewHarness(t)
	handler := h.Handler()

	w := testutil.DoRequest(t, handler, http.MethodGet, "/api/admin/posts", nil, h.SyncKey)
	if w.Code != http.StatusForbidden {
		t.Fatalf("sync key on admin: status=%d", w.Code)
	}

	w = testutil.DoRequest(t, handler, http.MethodPost, "/api/sync",
		testutil.SyncJSON("e2e-auth", "T", ""), h.AdminKey)
	if w.Code != http.StatusForbidden {
		t.Fatalf("admin key on sync: status=%d", w.Code)
	}
}

func TestE2E_syncRequiresAuth(t *testing.T) {
	h := NewHarness(t)
	w := testutil.DoRequest(t, h.Handler(), http.MethodPost, "/api/sync",
		testutil.SyncJSON("no-key", "T", ""), "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}
