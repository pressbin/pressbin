package parser_test

import (
	"strings"
	"testing"

	"pressbin.dev/pressbin/internal/parser"
)

func TestParse_frontMatterAndMarkdown(t *testing.T) {
	md := `---
title: Hello World
date: 2026-05-01
tags: [go, blog]
summary: Short intro
status: draft
---

# Hello

**Bold** text.
`
	got, err := parser.Parse([]byte(md))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Title != "Hello World" {
		t.Errorf("title = %q, want Hello World", got.Title)
	}
	if got.PublishedAt != "2026-05-01" {
		t.Errorf("date = %q", got.PublishedAt)
	}
	if got.Summary != "Short intro" {
		t.Errorf("summary = %q", got.Summary)
	}
	if got.Status != "draft" {
		t.Errorf("status = %q", got.Status)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "go" || got.Tags[1] != "blog" {
		t.Errorf("tags = %v", got.Tags)
	}
	if !strings.Contains(got.HTML, "<strong>Bold</strong>") {
		t.Errorf("HTML missing bold: %s", got.HTML)
	}
	if !strings.Contains(got.HTML, "<h1>Hello</h1>") {
		t.Errorf("HTML missing h1: %s", got.HTML)
	}
}

func TestParse_tagsAsCommaString(t *testing.T) {
	md := `---
title: T
tags: alpha, beta , gamma
---
body
`
	got, err := parser.Parse([]byte(md))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha", "beta", "gamma"}
	if len(got.Tags) != len(want) {
		t.Fatalf("tags = %v", got.Tags)
	}
	for i, w := range want {
		if got.Tags[i] != w {
			t.Errorf("tags[%d] = %q, want %q", i, got.Tags[i], w)
		}
	}
}

func TestParse_noFrontMatter(t *testing.T) {
	got, err := parser.Parse([]byte("# Only body\n\nParagraph."))
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "" {
		t.Errorf("title = %q", got.Title)
	}
	if !strings.Contains(got.HTML, "Only body") {
		t.Errorf("HTML = %s", got.HTML)
	}
}

func TestParse_invalidMarkdownStillRenders(t *testing.T) {
	_, err := parser.Parse([]byte("plain text without front matter"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
