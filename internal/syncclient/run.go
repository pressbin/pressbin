package syncclient

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Run syncs posts and assets under repo to Pressbin (incremental git diff on push).
func Run(c *Client, repo string) error {
	repo, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(repo, ".git")); err != nil {
		return fmt.Errorf("not a git repository: %s", repo)
	}

	if !hasParentCommit(repo) {
		fmt.Println("First push: syncing all posts and images")
		return syncAll(c, repo)
	}

	for _, f := range gitDiffNames(repo, true, "posts/*.md", "posts/**/*.md") {
		fmt.Printf("Deleting post %s...\n", f)
		if err := c.DeletePost(slugFromDeletedPost(repo, f)); err != nil {
			return err
		}
	}

	for _, f := range gitDiffNames(repo, false, "posts/*.md", "posts/**/*.md") {
		if !fileExists(filepath.Join(repo, f)) {
			continue
		}
		fmt.Printf("Syncing post %s...\n", f)
		if err := c.PushPost(filepath.Join(repo, f)); err != nil {
			return err
		}
	}

	for _, f := range gitDiffNames(repo, true, "assets/images/*", "assets/images/**/*") {
		fmt.Printf("Deleting asset %s...\n", f)
		if err := c.DeleteAsset(f); err != nil {
			return err
		}
	}

	for _, f := range gitDiffNames(repo, false, "assets/images/*", "assets/images/**/*") {
		if !fileExists(filepath.Join(repo, f)) {
			continue
		}
		fmt.Printf("Syncing asset %s...\n", f)
		if err := c.PushAsset(filepath.Join(repo, f)); err != nil {
			return err
		}
	}

	return nil
}

// RunAll uploads every posts/**/*.md and assets/images/** file (full resync, no git diff).
func RunAll(c *Client, repo string) error {
	repo, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	fmt.Println("Full sync: uploading all posts and images")
	return syncAll(c, repo)
}

func syncAll(c *Client, repo string) error {
	posts, err := findFiles(repo, "posts", ".md")
	if err != nil {
		return err
	}
	for _, f := range posts {
		fmt.Printf("Syncing post %s...\n", f)
		if err := c.PushPost(filepath.Join(repo, f)); err != nil {
			return err
		}
	}
	assets, err := findFilesUnder(repo, filepath.Join("assets", "images"))
	if err != nil {
		return err
	}
	for _, f := range assets {
		fmt.Printf("Syncing asset %s...\n", f)
		if err := c.PushAsset(filepath.Join(repo, f)); err != nil {
			return err
		}
	}
	return nil
}

func hasParentCommit(repo string) bool {
	cmd := exec.Command("git", "-C", repo, "rev-parse", "--verify", "HEAD~1")
	return cmd.Run() == nil
}

func gitRun(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	return cmd.Output()
}

func gitDiffNames(repo string, deleted bool, pathspecs ...string) []string {
	args := []string{"diff", "--name-only", "HEAD~1", "HEAD", "--"}
	if deleted {
		args = []string{"diff", "--diff-filter=D", "--name-only", "HEAD~1", "HEAD", "--"}
	}
	args = append(args, pathspecs...)
	out, err := gitRun(repo, args...)
	if err != nil {
		return nil
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}
	return names
}

func findFiles(repo, dir, ext string) ([]string, error) {
	root := filepath.Join(repo, dir)
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ext) {
			rel, err := filepath.Rel(repo, path)
			if err != nil {
				return err
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return out, nil
}

func findFilesUnder(repo, dir string) ([]string, error) {
	root := filepath.Join(repo, dir)
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(repo, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return out, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
