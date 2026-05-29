package syncclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type versionResponse struct {
	Version string `json:"version"`
}

// FetchVersion returns the server version from GET /api/version (GitHub Actions uses this).
func FetchVersion(baseURL string, client *http.Client) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("base URL is required")
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/version", nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("version: %d — %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var v versionResponse
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}
	if v.Version == "" {
		return "", fmt.Errorf("version: empty version in response")
	}
	return v.Version, nil
}
