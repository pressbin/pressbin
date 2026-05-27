package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type ParseResult struct {
	Slug        string
	Title       string
	Summary     string
	Tags        []string
	PublishedAt string
	Status      string
	HTML        string
}

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
		meta.Meta,
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

func Parse(content []byte) (ParseResult, error) {
	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := md.Convert(content, &buf, parser.WithContext(ctx)); err != nil {
		return ParseResult{}, err
	}

	metaData := meta.Get(ctx)

	return ParseResult{
		Slug:        getString(metaData, "slug"),
		Title:       getString(metaData, "title"),
		Summary:     getString(metaData, "summary"),
		Tags:        getStringSlice(metaData, "tags"),
		PublishedAt: getString(metaData, "date"),
		Status:      getString(metaData, "status"),
		HTML:        buf.String(),
	}, nil
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func getStringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			out = append(out, strings.TrimSpace(fmt.Sprint(x)))
		}
		return out
	case string:
		parts := strings.Split(t, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	default:
		return []string{strings.TrimSpace(fmt.Sprint(t))}
	}
}
