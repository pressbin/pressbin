package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"pressbin.dev/pressbin/internal/render"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	page := queryPage(r)
	limit := s.limit()
	posts, total, err := s.store.ListPosts(page, limit)
	if err != nil {
		writeError(w, 500, "failed to list posts")
		return
	}
	pd := s.pageData()
	html := render.IndexPage(posts, pd, page, totalPages(total, limit))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = html.Render(w)
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	post, err := s.store.GetPost(slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 404, "not found")
			return
		}
		writeError(w, 500, "failed to load post")
		return
	}
	if post.Status != "published" {
		writeError(w, 404, "not found")
		return
	}
	pd := s.pageData()
	html := render.PostPage(post, pd)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = html.Render(w)
}

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.store.AllTags()
	if err != nil {
		writeError(w, 500, "failed to list tags")
		return
	}
	pd := s.pageData()
	html := render.TagsPage(tags, pd)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = html.Render(w)
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	page := queryPage(r)
	limit := s.limit()
	posts, total, err := s.store.GetPostsByTag(tag, page, limit)
	if err != nil {
		writeError(w, 500, "failed to list posts")
		return
	}
	pd := s.pageData()
	html := render.TagPage(tag, posts, pd, page, totalPages(total, limit))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = html.Render(w)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	posts, err := s.store.SearchPosts(q)
	if err != nil {
		writeError(w, 500, "search failed")
		return
	}
	html := render.SearchResults(posts, q)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = html.Render(w)
}
