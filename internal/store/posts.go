package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Post struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	ContentMD   string    `json:"content_md"`
	ContentHTML string    `json:"content_html"`
	Summary     string    `json:"summary"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags"`
}

func (s *Store) UpsertPost(p Post) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if p.Status == "" {
		p.Status = "published"
	}

	_, err = tx.Exec(`
		INSERT INTO posts (slug, title, content_md, content_html, summary, published_at, updated_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			title = excluded.title,
			content_md = excluded.content_md,
			content_html = excluded.content_html,
			summary = excluded.summary,
			published_at = excluded.published_at,
			updated_at = excluded.updated_at,
			status = excluded.status
	`, p.Slug, p.Title, p.ContentMD, p.ContentHTML, p.Summary, p.PublishedAt, p.UpdatedAt, p.Status)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM tags WHERE slug = ?`, p.Slug); err != nil {
		return err
	}
	for _, t := range p.Tags {
		tag := strings.TrimSpace(strings.ToLower(t))
		if tag == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO tags (slug, tag) VALUES (?, ?)`, p.Slug, tag); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) PostExists(slug string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM posts WHERE slug = ?`, slug).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Store) ListPosts(page, limit int) ([]Post, int, error) {
	return s.listPostsWhere(`status = 'published'`, page, limit)
}

func (s *Store) ListAllPosts(page, limit int) ([]Post, int, error) {
	return s.listPostsWhere(`1=1`, page, limit)
}

func (s *Store) listPostsWhere(where string, page, limit int) ([]Post, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int
	qCount := fmt.Sprintf(`SELECT COUNT(1) FROM posts WHERE %s`, where)
	if err := s.db.QueryRow(qCount).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf(`
		SELECT slug, title, content_md, content_html, summary, published_at, updated_at, status
		FROM posts WHERE %s
		ORDER BY published_at DESC, slug ASC
		LIMIT ? OFFSET ?
	`, where)
	rows, err := s.db.Query(q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, 0, err
		}
		tags, err := s.tagsForSlug(p.Slug)
		if err != nil {
			return nil, 0, err
		}
		p.Tags = tags
		posts = append(posts, p)
	}
	return posts, total, rows.Err()
}

func scanPost(sc interface {
	Scan(dest ...any) error
}) (Post, error) {
	var p Post
	err := sc.Scan(
		&p.Slug, &p.Title, &p.ContentMD, &p.ContentHTML, &p.Summary,
		&p.PublishedAt, &p.UpdatedAt, &p.Status,
	)
	return p, err
}

func (s *Store) GetPost(slug string) (Post, error) {
	row := s.db.QueryRow(`
		SELECT slug, title, content_md, content_html, summary, published_at, updated_at, status
		FROM posts WHERE slug = ?
	`, slug)
	p, err := scanPost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, err
	}
	if err != nil {
		return Post{}, err
	}
	p.Tags, err = s.tagsForSlug(slug)
	return p, err
}

func (s *Store) GetPostsByTag(tag string, page, limit int) ([]Post, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	t := strings.TrimSpace(strings.ToLower(tag))

	var total int
	err := s.db.QueryRow(`
		SELECT COUNT(1) FROM posts p
		INNER JOIN tags t ON t.slug = p.slug
		WHERE t.tag = ? AND p.status = 'published'
	`, t).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(`
		SELECT p.slug, p.title, p.content_md, p.content_html, p.summary, p.published_at, p.updated_at, p.status
		FROM posts p
		INNER JOIN tags t ON t.slug = p.slug
		WHERE t.tag = ? AND p.status = 'published'
		ORDER BY p.published_at DESC, p.slug ASC
		LIMIT ? OFFSET ?
	`, t, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, 0, err
		}
		p.Tags, err = s.tagsForSlug(p.Slug)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}
	return posts, total, rows.Err()
}

func (s *Store) DeletePost(slug string) error {
	_, err := s.db.Exec(`DELETE FROM posts WHERE slug = ?`, slug)
	return err
}

func (s *Store) SetPostStatus(slug, status string) error {
	if status != "published" && status != "draft" {
		return fmt.Errorf("invalid status")
	}
	res, err := s.db.Exec(`UPDATE posts SET updated_at = ?, status = ? WHERE slug = ?`, time.Now().UTC(), status, slug)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// buildFTSQuery turns user input into an FTS5 prefix query so partial terms
// like "po" match "post" as the user types.
func buildFTSQuery(query string) string {
	terms := strings.Fields(strings.TrimSpace(query))
	if len(terms) == 0 {
		return ""
	}
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSuffix(term, "*")
		if term == "" {
			continue
		}
		term = strings.ReplaceAll(term, `"`, `""`)
		parts = append(parts, `"`+term+`"*`)
	}
	return strings.Join(parts, " ")
}

func (s *Store) SearchPosts(query string) ([]Post, error) {
	ftsQuery := buildFTSQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}
	rows, err := s.db.Query(`
		SELECT p.slug, p.title, p.content_md, p.content_html, p.summary, p.published_at, p.updated_at, p.status
		FROM fts_index
		INNER JOIN posts p ON p.rowid = fts_index.rowid
		WHERE fts_index MATCH ? AND p.status = 'published'
		ORDER BY rank
		LIMIT 50
	`, ftsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		p.Tags, err = s.tagsForSlug(p.Slug)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (s *Store) CountPostsByStatus() (total, published, draft int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(1) FROM posts`).Scan(&total)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COUNT(1) FROM posts WHERE status = 'published'`).Scan(&published)
	if err != nil {
		return
	}
	err = s.db.QueryRow(`SELECT COUNT(1) FROM posts WHERE status = 'draft'`).Scan(&draft)
	return
}

func (s *Store) CountDistinctTags() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(DISTINCT tag) FROM tags`).Scan(&n)
	return n, err
}
