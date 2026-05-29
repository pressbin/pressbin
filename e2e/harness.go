package e2e

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/server"
	"pressbin.dev/pressbin/internal/store"
	"pressbin.dev/pressbin/internal/syncclient"
	"pressbin.dev/pressbin/internal/testutil"
)

const defaultVersion = "v1.0.0-e2e"

// Harness wires a real Pressbin HTTP server with SQLite, theme assets, and optional blog assets dir.
type Harness struct {
	T         *testing.T
	Store     *store.Store
	Server    *httptest.Server
	SyncKey   string
	AdminKey  string
	AssetsDir string
	Version   string
}

// NewHarness starts an in-process Pressbin stack for cross-component tests.
func NewHarness(t *testing.T) *Harness {
	t.Helper()
	return NewHarnessWithVersion(t, defaultVersion)
}

// NewHarnessWithVersion is like NewHarness but sets GET /api/version.
func NewHarnessWithVersion(t *testing.T, version string) *Harness {
	t.Helper()

	st := testutil.NewStore(t)
	syncKey := testutil.MustSyncKey(t, st)
	adminKey := testutil.MustAdminKey(t, st)

	assetsDir := filepath.Join(t.TempDir(), "assets")
	cfg := testutil.TestConfig()
	cfg.Assets.Path = assetsDir

	srv := server.New(st, cfg, testutil.TestAssets(), version)
	ts := httptest.NewServer(srv.Handler())

	t.Cleanup(func() {
		ts.Close()
	})

	return &Harness{
		T:         t,
		Store:     st,
		Server:    ts,
		SyncKey:   syncKey,
		AdminKey:  adminKey,
		AssetsDir: assetsDir,
		Version:   version,
	}
}

// BaseURL is the public URL of the test instance (like PRESSBIN_URL in CI).
func (h *Harness) BaseURL() string {
	return h.Server.URL
}

// Handler returns the HTTP handler (for testutil.DoRequest).
func (h *Harness) Handler() http.Handler {
	return h.Server.Config.Handler
}

// SyncClient returns a sync client pointed at this harness (like pressbin-sync in Actions).
func (h *Harness) SyncClient() *syncclient.Client {
	return &syncclient.Client{
		BaseURL: h.BaseURL(),
		Key:     h.SyncKey,
		HTTP:    h.Server.Client(),
	}
}

// WithPressbinEnv sets PRESSBIN_URL and PRESSBIN_KEY for syncclient.NewFromEnv.
func (h *Harness) WithPressbinEnv() {
	h.T.Helper()
	h.T.Setenv("PRESSBIN_URL", h.BaseURL())
	h.T.Setenv("PRESSBIN_KEY", h.SyncKey)
}
