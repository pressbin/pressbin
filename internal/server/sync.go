package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"pressbin.dev/pressbin/internal/parser"
	"pressbin.dev/pressbin/internal/store"
)

type syncPayload struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Date    string   `json:"date"`
	Tags    []string `json:"tags"`
	Summary string   `json:"summary"`
	Content string   `json:"content"`
	Status  string   `json:"status"`
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	var req syncPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.Slug == "" || req.Content == "" {
		writeError(w, 400, "slug and content are required")
		return
	}

	exists, err := s.store.PostExists(req.Slug)
	if err != nil {
		writeError(w, 500, "failed to save post")
		return
	}

	parsed, err := parser.Parse([]byte(req.Content))
	if err != nil {
		writeError(w, 422, "failed to parse markdown")
		return
	}

	title := req.Title
	if title == "" {
		title = parsed.Title
	}
	if title == "" {
		title = req.Slug
	}
	tags := req.Tags
	if len(tags) == 0 && len(parsed.Tags) > 0 {
		tags = parsed.Tags
	}
	summary := req.Summary
	if summary == "" {
		summary = parsed.Summary
	}
	status := req.Status
	if status == "" {
		if parsed.Status != "" {
			status = parsed.Status
		} else {
			status = "published"
		}
	}

	pub := parseDate(req.Date)
	if req.Date == "" && parsed.PublishedAt != "" {
		pub = parseDate(parsed.PublishedAt)
	}

	post := store.Post{
		Slug:        req.Slug,
		Title:       title,
		ContentMD:   req.Content,
		ContentHTML: parsed.HTML,
		Summary:     summary,
		Tags:        tags,
		Status:      status,
		PublishedAt: pub,
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.store.UpsertPost(post); err != nil {
		writeError(w, 500, "failed to save post")
		return
	}

	_ = s.store.SetSetting("last_sync_at", time.Now().UTC().Format(time.RFC3339))

	action := "updated"
	if !exists {
		action = "created"
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "slug": req.Slug, "action": action})
}

func (s *Server) handleSyncDelete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		writeError(w, 400, "slug is required")
		return
	}
	exists, err := s.store.PostExists(slug)
	if err != nil {
		writeError(w, 500, "failed to delete post")
		return
	}
	if !exists {
		writeError(w, 404, "not found")
		return
	}
	if err := s.store.DeletePost(slug); err != nil {
		writeError(w, 500, "failed to delete post")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "slug": slug, "action": "deleted"})
}
