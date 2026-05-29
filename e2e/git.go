package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// InitGitRepo creates a new git repository under t.TempDir().
func InitGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	Git(t, dir, "init")
	Git(t, dir, "config", "user.email", "e2e@pressbin.dev")
	Git(t, dir, "config", "user.name", "Pressbin E2E")
	return dir
}

// WriteRepoFile writes content at repo-relative path.
func WriteRepoFile(t *testing.T, repo, rel, content string) {
	t.Helper()
	path := filepath.Join(repo, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// GitCommitAll stages and commits all changes in repo.
func GitCommitAll(t *testing.T, repo, message string) {
	t.Helper()
	Git(t, repo, "add", "-A")
	Git(t, repo, "commit", "-m", message)
}

// Git runs a git command in repo.
func Git(t *testing.T, repo string, args ...string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Pressbin E2E",
		"GIT_AUTHOR_EMAIL=e2e@pressbin.dev",
		"GIT_COMMITTER_NAME=Pressbin E2E",
		"GIT_COMMITTER_EMAIL=e2e@pressbin.dev",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
