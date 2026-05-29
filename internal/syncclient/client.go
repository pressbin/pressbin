package syncclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pressbin.dev/pressbin/internal/parser"
)

// Client pushes content-repo changes to a Pressbin instance.
type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewFromEnv() (*Client, error) {
	base := strings.TrimSpace(os.Getenv("PRESSBIN_URL"))
	key := strings.TrimSpace(os.Getenv("PRESSBIN_KEY"))
	if base == "" {
		return nil, fmt.Errorf("PRESSBIN_URL is required")
	}
	if key == "" {
		return nil, fmt.Errorf("PRESSBIN_KEY is required")
	}
	return &Client{
		BaseURL: strings.TrimRight(base, "/"),
		Key:     key,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

type postPayload struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags"`
	Summary string   `json:"summary"`
	Content string   `json:"content"`
	Status  string   `json:"status"`
}

type assetPayload struct {
	Path          string `json:"path"`
	ContentBase64 string `json:"content_base64"`
}

func (c *Client) PushPost(filePath string) error {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	parsed, err := parser.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filePath, err)
	}
	slug := parsed.Slug
	if slug == "" {
		slug = strings.TrimSuffix(filepath.Base(filePath), ".md")
	}
	title := parsed.Title
	if title == "" {
		title = slug
	}
	status := parsed.Status
	if status == "" {
		status = "published"
	}
	payload := postPayload{
		Slug:    slug,
		Title:   title,
		Date:    parsed.PublishedAt,
		Tags:    parsed.Tags,
		Summary: parsed.Summary,
		Content: string(raw),
		Status:  status,
	}
	var resp map[string]string
	if err := c.doJSON(http.MethodPost, "/api/sync", payload, &resp); err != nil {
		return err
	}
	fmt.Printf("OK: %s\n", slug)
	return nil
}

func (c *Client) DeletePost(slug string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/sync/"+slug, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete %s: %d — %s", slug, res.StatusCode, strings.TrimSpace(string(body)))
	}
	fmt.Printf("OK: deleted %s\n", slug)
	return nil
}

func (c *Client) PushAsset(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	rel, err := assetRelPath(filePath)
	if err != nil {
		return err
	}
	payload := assetPayload{
		Path:          rel,
		ContentBase64: base64.StdEncoding.EncodeToString(data),
	}
	var resp map[string]string
	if err := c.doJSON(http.MethodPost, "/api/sync/asset", payload, &resp); err != nil {
		return err
	}
	fmt.Printf("OK: %s\n", rel)
	return nil
}

func (c *Client) DeleteAsset(filePath string) error {
	rel, err := assetRelPath(filePath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/sync/asset/"+rel, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete asset %s: %d — %s", rel, res.StatusCode, strings.TrimSpace(string(body)))
	}
	fmt.Printf("OK: deleted %s\n", rel)
	return nil
}

func (c *Client) doJSON(method, path string, payload, out any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: %d — %s", method, path, res.StatusCode, strings.TrimSpace(string(body)))
	}
	if out != nil {
		_ = json.Unmarshal(body, out)
	}
	return nil
}

func assetRelPath(filePath string) (string, error) {
	p := filepath.ToSlash(filePath)
	if i := strings.Index(p, "assets/"); i >= 0 {
		p = p[i+len("assets/"):]
	}
	if p == "" || strings.Contains(p, "..") {
		return "", fmt.Errorf("invalid asset path: %s", filePath)
	}
	return p, nil
}

func slugFromDeletedPost(repo, filePath string) string {
	out, err := gitRun(repo, "show", "HEAD~1:"+filePath)
	if err == nil {
		if parsed, err := parser.Parse([]byte(out)); err == nil && parsed.Slug != "" {
			return parsed.Slug
		}
	}
	return strings.TrimSuffix(filepath.Base(filePath), ".md")
}
