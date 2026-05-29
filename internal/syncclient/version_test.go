package syncclient_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pressbin.dev/pressbin/internal/syncclient"
)

func TestFetchVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"version":"v9.9.9"}`))
	}))
	defer srv.Close()

	v, err := syncclient.FetchVersion(srv.URL, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if v != "v9.9.9" {
		t.Fatalf("version = %q", v)
	}
}
