package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
)

const maxAssetBytes = 5 << 20 // 5 MiB

var errAssetsNotConfigured = errors.New("assets upload storage_path is not configured")

type assetSyncPayload struct {
	Path          string `json:"path"`
	ContentBase64 string `json:"content_base64"`
}

// normalizeAssetPath returns a path relative to assets.upload.storage_path
// (e.g. images/foo.svg, fonts/foo.woff2).
func (s *Server) normalizeAssetPath(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "/")
	for strings.HasPrefix(p, "assets/") {
		p = strings.TrimPrefix(p, "assets/")
	}
	p = path.Clean(p)
	if p == "." || p == ".." || strings.HasPrefix(p, "../") || strings.Contains(p, "..") {
		return "", fmt.Errorf("invalid asset path")
	}

	allowed := s.config.Assets.AllowedPrefixes
	if len(allowed) == 0 {
		allowed = []string{"images/", "fonts/", "downloads/", "css/"}
	}
	ok := false
	for _, pref := range allowed {
		pref = strings.TrimSpace(pref)
		if pref == "" {
			continue
		}
		if !strings.HasSuffix(pref, "/") {
			pref += "/"
		}
		if strings.HasPrefix(p, pref) {
			ok = true
			break
		}
	}
	if !ok {
		return "", fmt.Errorf("asset path is not allowed")
	}

	base := path.Base(p)
	if base == "" || strings.HasPrefix(base, ".") {
		return "", fmt.Errorf("invalid asset path")
	}
	return p, nil
}

func (s *Server) handleSyncAsset(w http.ResponseWriter, r *http.Request) {
	var data []byte
	rel := ""

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(maxAssetBytes + 1024); err != nil {
			writeError(w, 400, "invalid multipart form")
			return
		}
		rel = r.FormValue("path")
		file, _, err := r.FormFile("file")
		if err != nil {
			writeError(w, 400, "file is required")
			return
		}
		defer file.Close()
		data, err = io.ReadAll(io.LimitReader(file, maxAssetBytes+1))
		if err != nil {
			writeError(w, 400, "failed to read file")
			return
		}
	} else {
		var req assetSyncPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, 400, "invalid request body")
			return
		}
		rel = req.Path
		if req.ContentBase64 == "" {
			writeError(w, 400, "path and content_base64 are required")
			return
		}
		var err error
		data, err = base64.StdEncoding.DecodeString(req.ContentBase64)
		if err != nil {
			writeError(w, 400, "invalid content_base64")
			return
		}
	}

	rel, err := s.normalizeAssetPath(rel)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if len(data) == 0 {
		writeError(w, 400, "empty asset")
		return
	}
	if len(data) > maxAssetBytes {
		writeError(w, 413, "asset too large")
		return
	}

	exists := false
	if s.uploader != nil {
		if ok, err := s.uploader.Exists(rel); err == nil && ok {
			exists = true
		}
	}

	if s.uploader == nil {
		writeError(w, 503, "assets upload storage_path is not configured on the server")
		return
	}

	if err := s.uploader.Upload(rel, data); err != nil {
		if errors.Is(err, errAssetsNotConfigured) {
			writeError(w, 503, "assets upload storage_path is not configured on the server")
			return
		}
		writeError(w, 500, "failed to save asset")
		return
	}

	action := "updated"
	if !exists {
		action = "created"
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "path": rel, "action": action})
}

func (s *Server) handleSyncDeleteAsset(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	rel, err := s.normalizeAssetPath(raw)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if s.uploader == nil {
		writeError(w, 503, "assets upload storage_path is not configured on the server")
		return
	}
	if err := s.uploader.Delete(rel); err != nil {
		if errors.Is(err, errAssetsNotConfigured) {
			writeError(w, 503, "assets upload storage_path is not configured on the server")
			return
		}
		if os.IsNotExist(err) {
			writeError(w, 404, "not found")
			return
		}
		writeError(w, 500, "failed to delete asset")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "path": rel, "action": "deleted"})
}
