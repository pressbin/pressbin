package setup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pressbin.dev/pressbin/internal/setup"
)

func TestRun_createsLayoutAndKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	gotHome, err := setup.DefaultHome()
	if err != nil {
		t.Fatal(err)
	}
	if gotHome != filepath.Join(home, ".pressbin") {
		t.Fatalf("DefaultHome = %q", gotHome)
	}

	res, err := setup.Run(setup.Options{
		Home:      gotHome,
		SiteURL:   "https://blog.example.com",
		SiteTitle: "Test Blog",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{
		res.Layout.Config,
		res.Layout.DB,
		res.Layout.Assets,
		res.Layout.AdminKeyFile,
		res.Layout.SyncKeyFile,
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	admin, err := os.ReadFile(res.Layout.AdminKeyFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(admin), "pb_admin_") {
		t.Errorf("admin key = %q", admin)
	}
	sync, err := os.ReadFile(res.Layout.SyncKeyFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(sync), "pb_sync_") {
		t.Errorf("sync key = %q", sync)
	}

	_, err = setup.Run(setup.Options{
		Home:    gotHome,
		SiteURL: "https://blog.example.com",
	})
	if err != nil {
		t.Fatalf("second setup: %v", err)
	}
}

func TestRun_requiresSiteURL(t *testing.T) {
	_, err := setup.Run(setup.Options{Home: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "site-url") {
		t.Fatalf("err = %v", err)
	}
}
