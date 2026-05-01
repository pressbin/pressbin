package store

import (
	"sort"
)

func (s *Store) tagsForSlug(slug string) ([]string, error) {
	rows, err := s.db.Query(`SELECT tag FROM tags WHERE slug = ? ORDER BY tag`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *Store) AllTags() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT tag FROM tags ORDER BY tag`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(tags)
	return tags, nil
}
