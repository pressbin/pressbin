package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/render"
	"pressbin.dev/pressbin/internal/store"
)

type Server struct {
	store  *store.Store
	config *config.Config
	router chi.Router
	assets fs.FS
}

func New(st *store.Store, cfg *config.Config, assets fs.FS) *Server {
	s := &Server{store: st, config: cfg, assets: assets}
	s.router = chi.NewRouter()
	s.routes()
	return s
}

func (s *Server) routes() {
	r := s.router

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	sub, err := fs.Sub(s.assets, "assets")
	if err != nil {
		slog.Error("assets sub fs", "err", err)
		sub = s.assets
	}
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.FS(sub))))

	r.Get("/", s.handleIndex)
	r.Get("/p/{slug}", s.handlePost)
	r.Get("/tags", s.handleTags)
	r.Get("/tag/{tag}", s.handleTag)
	r.Get("/search", s.handleSearch)
	r.Get("/feed.xml", s.handleRSS)

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth("posts:write"))
		r.Post("/api/sync", s.handleSync)
		r.Delete("/api/sync/{slug}", s.handleSyncDelete)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth("*"))
		r.Get("/api/admin/posts", s.handleAdminListPosts)
		r.Get("/api/admin/posts/{slug}", s.handleAdminGetPost)
		r.Put("/api/admin/posts/{slug}", s.handleAdminUpdatePost)
		r.Delete("/api/admin/posts/{slug}", s.handleAdminDeletePost)
		r.Patch("/api/admin/posts/{slug}/status", s.handleAdminSetStatus)

		r.Get("/api/admin/keys", s.handleAdminListKeys)
		r.Post("/api/admin/keys", s.handleAdminCreateKey)
		r.Delete("/api/admin/keys/{id}", s.handleAdminRevokeKey)

		r.Get("/api/admin/settings", s.handleAdminGetSettings)
		r.Put("/api/admin/settings", s.handleAdminUpdateSettings)

		r.Get("/api/admin/stats", s.handleAdminStats)
	})
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func queryPage(r *http.Request) int {
	q := r.URL.Query().Get("page")
	if q == "" {
		return 1
	}
	n, err := strconv.Atoi(q)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

func (s *Server) limit() int {
	n := s.store.PostsPerPageFromSettings(s.config.Site.PostsPerPage)
	if n < 1 {
		return 10
	}
	return n
}

func (s *Server) pageData() render.PageData {
	m, err := s.store.GetAllSettings()
	if err != nil {
		m = map[string]string{}
	}
	get := func(key, fallback string) string {
		if v := strings.TrimSpace(m[key]); v != "" {
			return v
		}
		return fallback
	}
	return render.PageData{
		SiteTitle:       get("site_title", s.config.Site.Title),
		SiteDescription: get("site_description", s.config.Site.Description),
		SiteURL:         get("site_url", s.config.Site.URL),
	}
}
