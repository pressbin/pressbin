package testutil

import "encoding/json"

// SyncJSON builds a POST /api/sync request body with markdown content.
func SyncJSON(slug, title, markdown string) []byte {
	if markdown == "" {
		markdown = "---\ntitle: " + title + "\ndate: 2026-05-01\ntags: [sync, test]\n---\n\n# " + title + "\n\nBody.\n"
	}
	raw, _ := json.Marshal(map[string]string{
		"slug":    slug,
		"title":   title,
		"content": markdown,
	})
	return raw
}
