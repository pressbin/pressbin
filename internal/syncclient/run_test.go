package syncclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/syncclient"
)

func TestRunAll_fullSync(t *testing.T) {
	var syncCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && (r.URL.Path == "/api/sync" || r.URL.Path == "/api/sync/asset") {
			syncCount++
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}
	}))
	defer srv.Close()

	repo := t.TempDir()
	writeFile(t, repo, "posts/a.md", "---\ntitle: A\nslug: a\n---\n\nx\n")
	writeFile(t, repo, "assets/images/x.svg", "<svg/>")

	c := &syncclient.Client{BaseURL: srv.URL, Key: "k", HTTP: srv.Client()}
	if err := syncclient.RunAll(c, repo); err != nil {
		t.Fatalf("RunAll: %v", err)
	}
	if syncCount != 2 {
		t.Fatalf("sync calls = %d, want 2", syncCount)
	}
}

func TestRun_firstPush(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	var syncCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && (r.URL.Path == "/api/sync" || r.URL.Path == "/api/sync/asset") {
			syncCount++
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}
	}))
	defer srv.Close()

	repo := initGitRepo(t)
	writeFile(t, repo, "posts/a.md", "---\ntitle: A\nslug: a\n---\n\nx\n")
	writeFile(t, repo, "assets/images/x.svg", "<svg/>")
	gitCommitAll(t, repo, "init")

	t.Setenv("PRESSBIN_URL", srv.URL)
	t.Setenv("PRESSBIN_KEY", "k")
	c := &syncclient.Client{BaseURL: srv.URL, Key: "k", HTTP: srv.Client()}
	if err := syncclient.Run(c, repo); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if syncCount != 2 {
		t.Fatalf("sync calls = %d, want 2", syncCount)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	return dir
}

func gitCommitAll(t *testing.T, dir, msg string) {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", msg)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=t@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, repo, rel, content string) {
	t.Helper()
	path := filepath.Join(repo, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
