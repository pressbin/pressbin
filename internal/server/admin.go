package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jaevor/go-nanoid"
	"golang.org/x/crypto/bcrypt"

	"pressbin.dev/pressbin/internal/parser"
	"pressbin.dev/pressbin/internal/store"
)

func (s *Server) handleAdminListPosts(w http.ResponseWriter, r *http.Request) {
	page := queryPage(r)
	limit := s.limit()
	posts, total, err := s.store.ListAllPosts(page, limit)
	if err != nil {
		writeError(w, 500, "failed to list posts")
		return
	}
	type row struct {
		Slug        string    `json:"slug"`
		Title       string    `json:"title"`
		Summary     string    `json:"summary"`
		Tags        []string  `json:"tags"`
		Status      string    `json:"status"`
		PublishedAt time.Time `json:"published_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	out := make([]row, 0, len(posts))
	for _, p := range posts {
		out = append(out, row{
			Slug: p.Slug, Title: p.Title, Summary: p.Summary, Tags: p.Tags,
			Status: p.Status, PublishedAt: p.PublishedAt, UpdatedAt: p.UpdatedAt,
		})
	}
	writeJSON(w, 200, map[string]any{"posts": out, "total": total, "page": page})
}

func (s *Server) handleAdminGetPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := s.store.GetPost(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 404, "not found")
			return
		}
		writeError(w, 500, "failed to load post")
		return
	}
	writeJSON(w, 200, p)
}

func (s *Server) handleAdminUpdatePost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var req syncPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, 400, "content is required")
		return
	}
	req.Slug = slug

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
		title = slug
	}
	tags := req.Tags
	if len(tags) == 0 {
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
	if req.Date == "" {
		if parsed.PublishedAt != "" {
			pub = parseDate(parsed.PublishedAt)
		}
		existing, gerr := s.store.GetPost(slug)
		if gerr == nil {
			pub = existing.PublishedAt
		} else if !errors.Is(gerr, sql.ErrNoRows) {
			writeError(w, 500, "failed to load post")
			return
		}
	}
	post := store.Post{
		Slug: slug, Title: title, ContentMD: req.Content, ContentHTML: parsed.HTML,
		Summary: summary, Tags: tags, Status: status, PublishedAt: pub, UpdatedAt: time.Now().UTC(),
	}
	if err := s.store.UpsertPost(post); err != nil {
		writeError(w, 500, "failed to save post")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "slug": slug})
}

func (s *Server) handleAdminDeletePost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := s.store.DeletePost(slug); err != nil {
		writeError(w, 500, "failed to delete")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminSetStatus(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if err := s.store.SetPostStatus(slug, body.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 404, "not found")
			return
		}
		writeError(w, 400, "invalid status")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.store.ListKeys()
	if err != nil {
		writeError(w, 500, "failed to list keys")
		return
	}
	type row struct {
		ID          string     `json:"id"`
		Label       string     `json:"label"`
		Prefix      string     `json:"prefix"`
		Permissions []string   `json:"permissions"`
		CreatedAt   time.Time  `json:"created_at"`
		LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	}
	out := make([]row, 0, len(keys))
	for _, k := range keys {
		out = append(out, row{
			ID: k.ID, Label: k.Label, Prefix: k.Prefix, Permissions: k.Permissions,
			CreatedAt: k.CreatedAt, LastUsedAt: k.LastUsedAt,
		})
	}
	writeJSON(w, 200, map[string]any{"keys": out})
}

func (s *Server) handleAdminCreateKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label       string   `json:"label"`
		Type        string   `json:"type"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	prefix, defaultPerms, err := store.KeyPrefixForType(body.Type)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	perms := body.Permissions
	if len(perms) == 0 {
		perms = defaultPerms
	}
	raw := prefix + store.RandomKeySuffix(32)
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, 500, "failed to create key")
		return
	}
	idGen, err := nanoid.Standard(21)
	if err != nil {
		writeError(w, 500, "failed to create key")
		return
	}
	k := store.APIKey{
		ID: idGen(), Label: body.Label, Prefix: prefix, Hash: string(hash),
		Permissions: perms, CreatedAt: time.Now().UTC(),
	}
	if err := s.store.CreateKey(k); err != nil {
		writeError(w, 500, "failed to save key")
		return
	}
	writeJSON(w, 200, map[string]string{
		"id": k.ID, "key": raw, "label": body.Label,
		"note": "Save this key. It will not be shown again.",
	})
}

func (s *Server) handleAdminRevokeKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.store.RevokeKey(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 404, "not found")
			return
		}
		writeError(w, 500, "failed to revoke")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) handleAdminGetSettings(w http.ResponseWriter, r *http.Request) {
	m, err := s.store.GetAllSettings()
	if err != nil {
		writeError(w, 500, "failed to load settings")
		return
	}
	get := func(k string) string { return m[k] }
	n := s.config.Site.PostsPerPage
	if v := get("posts_per_page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			n = parsed
		}
	}
	writeJSON(w, 200, map[string]any{
		"site_title":       get("site_title"),
		"site_description": get("site_description"),
		"site_url":         get("site_url"),
		"posts_per_page":   n,
	})
}

func (s *Server) handleAdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	updates := map[string]string{}
	if v, ok := body["site_title"].(string); ok {
		updates["site_title"] = v
	}
	if v, ok := body["site_description"].(string); ok {
		updates["site_description"] = v
	}
	if v, ok := body["site_url"].(string); ok {
		updates["site_url"] = v
	}
	if v, ok := body["posts_per_page"]; ok {
		switch t := v.(type) {
		case float64:
			updates["posts_per_page"] = strconv.Itoa(int(t))
		case int:
			updates["posts_per_page"] = strconv.Itoa(t)
		case string:
			updates["posts_per_page"] = t
		}
	}
	if err := s.store.UpdateSettings(updates); err != nil {
		writeError(w, 500, "failed to update settings")
		return
	}
	s.handleAdminGetSettings(w, r)
}

func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	total, published, draft, err := s.store.CountPostsByStatus()
	if err != nil {
		writeError(w, 500, "failed to load stats")
		return
	}
	tags, err := s.store.CountDistinctTags()
	if err != nil {
		writeError(w, 500, "failed to load stats")
		return
	}
	var lastSync *time.Time
	if raw, err := s.store.GetSetting("last_sync_at"); err == nil && raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			lastSync = &t
		}
	}
	if lastSync == nil {
		lastSync, _ = s.store.LastPostUpdate()
	}
	resp := map[string]any{
		"total_posts": total, "published_posts": published, "draft_posts": draft,
		"total_tags": tags,
	}
	if lastSync != nil {
		resp["last_sync_at"] = lastSync.UTC().Format(time.RFC3339)
	} else {
		resp["last_sync_at"] = nil
	}
	writeJSON(w, 200, resp)
}
