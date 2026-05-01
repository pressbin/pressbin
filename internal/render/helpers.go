package render

import (
	"html"
	"regexp"
	"strings"
)

var tagRE = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	s = tagRE.ReplaceAllString(s, " ")
	return html.UnescapeString(strings.TrimSpace(s))
}

func trimWords(s string, maxWords int) string {
	fields := strings.Fields(s)
	if len(fields) <= maxWords {
		return s
	}
	return strings.Join(fields[:maxWords], " ") + "…"
}
