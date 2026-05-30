package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	assetsupload "pressbin.dev/pressbin/internal/assets"
	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/render"
	"pressbin.dev/pressbin/internal/store"
)

type Server struct {
	store    *store.Store
	config   *config.Config
	version  string
	router   chi.Router
	assets   fs.FS
	uploader assetsupload.Uploader
}

func New(st *store.Store, cfg *config.Config, themeFS fs.FS, version string) *Server {
	var uploader assetsupload.Uploader
	if dir := strings.TrimSpace(cfg.Assets.Path); dir != "" {
		u, err := assetsupload.NewLocalUploader(dir)
		if err == nil {
			uploader = u
		} else {
			slog.Warn("assets.path uploader", "path", dir, "err", err)
		}
	}

	if version == "" {
		version = "dev"
	}
	s := &Server{store: st, config: cfg, version: version, assets: themeFS, uploader: uploader}
	s.router = chi.NewRouter()
	s.routes()
	return s
}

func (s *Server) routes() {
	r := s.router

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	if dir := strings.TrimSpace(s.config.Assets.Path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			slog.Warn("assets.path not writable", "path", dir, "err", err)
		} else if st, err := os.Stat(dir); err != nil {
			slog.Warn("assets.path not readable", "path", dir, "err", err)
		} else if !st.IsDir() {
			slog.Warn("assets.path is not a directory", "path", dir)
		} else {
			slog.Info("serving blog assets", "path", dir, "url", "/assets/")
			r.Handle("/assets/*", staticCache(http.StripPrefix("/assets/", http.FileServer(http.Dir(dir)))))
		}
	}

	if sub, err := fs.Sub(s.assets, "assets"); err == nil {
		r.Handle("/theme/*", staticCache(http.StripPrefix("/theme/", http.FileServer(http.FS(sub)))))
	} else {
		slog.Error("theme assets", "err", err)
	}

	r.Get("/", s.handleIndex)
	r.Get("/p/{slug}", s.handlePost)
	r.Get("/tags", s.handleTags)
	r.Get("/tag/{tag}", s.handleTag)
	r.Get("/search", s.handleSearch)
	r.Get("/feed.xml", s.handleRSS)
	r.Get("/api/version", s.handleVersion)

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth("posts:write"))
		r.Post("/api/sync", s.handleSync)
		r.Delete("/api/sync/{slug}", s.handleSyncDelete)
		r.Post("/api/sync/asset", s.handleSyncAsset)
		r.Delete("/api/sync/asset/*", s.handleSyncDeleteAsset)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth("admin"))
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
		CustomCSSURL:    s.config.Theme.CustomCSSURL,
	}
}
