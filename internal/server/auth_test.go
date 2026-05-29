package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"pressbin.dev/pressbin/internal/testutil"
)

func TestAuth_missingAndInvalidKey(t *testing.T) {
	st := testutil.NewStore(t)
	srv := testutil.NewServer(t, st)
	h := srv.Handler()

	w := testutil.DoRequest(t, h, http.MethodGet, "/api/admin/posts", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}

	w = testutil.DoRequest(t, h, http.MethodGet, "/api/admin/posts", nil, "pb_admin_invalid")
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAuth_syncKeyCannotAccessAdmin(t *testing.T) {
	st := testutil.NewStore(t)
	syncKey := testutil.MustSyncKey(t, st)
	srv := testutil.NewServer(t, st)

	w := testutil.DoRequest(t, srv.Handler(), http.MethodGet, "/api/admin/posts", nil, syncKey)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAuth_adminKeyCannotSync(t *testing.T) {
	st := testutil.NewStore(t)
	admin := testutil.MustAdminKey(t, st)
	srv := testutil.NewServer(t, st)
	h := srv.Handler()

	w := testutil.DoRequest(t, h, http.MethodPost, "/api/sync", testutil.SyncJSON("x", "Title", ""), admin)
	if w.Code != http.StatusForbidden {
		t.Fatalf("sync status=%d body=%s", w.Code, w.Body.String())
	}

	body, _ := json.Marshal(map[string]string{
		"path":           "images/x.svg",
		"content_base64": "PHN2Zy8+",
	})
	w = testutil.DoRequest(t, h, http.MethodPost, "/api/sync/asset", body, admin)
	if w.Code != http.StatusForbidden {
		t.Fatalf("sync asset status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuth_adminKeyAllowed(t *testing.T) {
	st := testutil.NewStore(t)
	admin := testutil.MustAdminKey(t, st)
	srv := testutil.NewServer(t, st)

	w := testutil.DoRequest(t, srv.Handler(), http.MethodGet, "/api/admin/posts", nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["posts"]; !ok {
		t.Errorf("response = %v", resp)
	}
}
