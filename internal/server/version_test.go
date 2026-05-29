package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"pressbin.dev/pressbin/internal/testutil"
)

func TestVersion_public(t *testing.T) {
	st := testutil.NewStore(t)
	srv := testutil.NewServerWithVersion(t, st, "v1.2.3")
	h := srv.Handler()

	w := testutil.DoRequest(t, h, http.MethodGet, "/api/version", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Version != "v1.2.3" {
		t.Fatalf("version = %q", body.Version)
	}
}
