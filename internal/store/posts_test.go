package store

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestUpsertPost_createUpdateAndTags(t *testing.T) {
	st := openTestStore(t)
	slug := "unit-test-post-" + RandomKeySuffix(8)
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	p := Post{
		Slug: slug, Title: "First", ContentMD: "md", ContentHTML: "<p>html</p>",
		Summary: "sum", Tags: []string{"Go", " TEST "}, Status: "published",
		PublishedAt: now, UpdatedAt: now,
	}
	if err := st.UpsertPost(p); err != nil {
		t.Fatal(err)
	}

	got, err := st.GetPost(slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "First" {
		t.Errorf("title = %q", got.Title)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "go" || got.Tags[1] != "test" {
		t.Errorf("normalized tags = %v", got.Tags)
	}

	p.Title = "Updated"
	p.Tags = []string{"new-tag"}
	if err := st.UpsertPost(p); err != nil {
		t.Fatal(err)
	}
	got, _ = st.GetPost(slug)
	if got.Title != "Updated" {
		t.Errorf("title after update = %q", got.Title)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "new-tag" {
		t.Errorf("tags after update = %v", got.Tags)
	}
}

func TestUpsertPost_defaultStatus(t *testing.T) {
	st := openTestStore(t)
	slug := "draft-default-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	p := Post{
		Slug: slug, Title: "T", ContentMD: "m", ContentHTML: "h",
		PublishedAt: now, UpdatedAt: now,
	}
	if err := st.UpsertPost(p); err != nil {
		t.Fatal(err)
	}
	got, _ := st.GetPost(slug)
	if got.Status != "published" {
		t.Errorf("status = %q", got.Status)
	}
}

func TestListPosts_excludesDrafts(t *testing.T) {
	st := openTestStore(t)
	pub := "list-pub-" + RandomKeySuffix(8)
	draft := "list-draft-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	for _, slug := range []string{pub, draft} {
		status := "published"
		if slug == draft {
			status = "draft"
		}
		if err := st.UpsertPost(Post{
			Slug: slug, Title: slug, ContentMD: "m", ContentHTML: "h",
			Status: status, PublishedAt: now, UpdatedAt: now,
		}); err != nil {
			t.Fatal(err)
		}
	}

	posts, _, err := st.ListPosts(1, 1000)
	if err != nil {
		t.Fatal(err)
	}
	foundPub, foundDraft := false, false
	for _, p := range posts {
		if p.Slug == pub {
			foundPub = true
		}
		if p.Slug == draft {
			foundDraft = true
		}
	}
	if !foundPub {
		t.Error("published post missing from ListPosts")
	}
	if foundDraft {
		t.Error("draft post should not appear in ListPosts")
	}
}

func TestListAllPosts_includesDrafts(t *testing.T) {
	st := openTestStore(t)
	slug := "all-draft-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "D", ContentMD: "m", ContentHTML: "h",
		Status: "draft", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	posts, _, err := st.ListAllPosts(1, 1000)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range posts {
		if p.Slug == slug {
			found = true
			break
		}
	}
	if !found {
		t.Error("draft missing from ListAllPosts")
	}
}

func TestGetPostsByTag(t *testing.T) {
	st := openTestStore(t)
	slug := "tag-post-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	tag := "unique-tag-" + RandomKeySuffix(6)
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "Tagged", ContentMD: "m", ContentHTML: "h",
		Tags: []string{tag}, Status: "published", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	posts, total, err := st.GetPostsByTag(tag, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total < 1 {
		t.Fatalf("total = %d", total)
	}
	found := false
	for _, p := range posts {
		if p.Slug == slug {
			found = true
		}
	}
	if !found {
		t.Error("post not found by tag")
	}
}

func TestSearchPosts(t *testing.T) {
	st := openTestStore(t)
	slug := "search-xyzzy-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	needle := "xyzzyneedle" + RandomKeySuffix(6)
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "Search me", ContentMD: needle, ContentHTML: "<p>" + needle + "</p>",
		Summary: "s", Status: "published", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	posts, err := st.SearchPosts(needle)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range posts {
		if p.Slug == slug {
			found = true
		}
	}
	if !found {
		t.Errorf("search did not find post with needle %q", needle)
	}
}

func TestSearchPosts_emptyQuery(t *testing.T) {
	st := openTestStore(t)
	posts, err := st.SearchPosts("   ")
	if err != nil {
		t.Fatal(err)
	}
	if posts != nil {
		t.Errorf("expected nil slice, got %v", posts)
	}
}

func TestSearchPosts_prefix(t *testing.T) {
	st := openTestStore(t)
	slug := "prefix-post-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "My First Post", ContentMD: "hello world", ContentHTML: "<p>hello</p>",
		Summary: "s", Status: "published", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"po", "fir", "my f"} {
		posts, err := st.SearchPosts(q)
		if err != nil {
			t.Fatalf("SearchPosts(%q): %v", q, err)
		}
		found := false
		for _, p := range posts {
			if p.Slug == slug {
				found = true
			}
		}
		if !found {
			t.Errorf("prefix search %q did not find post %q", q, slug)
		}
	}
}

func TestDeletePost_andPostExists(t *testing.T) {
	st := openTestStore(t)
	slug := "delete-me-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "D", ContentMD: "m", ContentHTML: "h",
		Status: "published", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	ok, err := st.PostExists(slug)
	if err != nil || !ok {
		t.Fatalf("PostExists: ok=%v err=%v", ok, err)
	}
	if err := st.DeletePost(slug); err != nil {
		t.Fatal(err)
	}
	ok, _ = st.PostExists(slug)
	if ok {
		t.Error("post still exists after delete")
	}
	_, err = st.GetPost(slug)
	if err != sql.ErrNoRows {
		t.Errorf("GetPost err = %v", err)
	}
}

func TestSetPostStatus(t *testing.T) {
	st := openTestStore(t)
	slug := "status-change-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "S", ContentMD: "m", ContentHTML: "h",
		Status: "published", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.SetPostStatus(slug, "draft"); err != nil {
		t.Fatal(err)
	}
	got, _ := st.GetPost(slug)
	if got.Status != "draft" {
		t.Errorf("status = %q", got.Status)
	}
	if err := st.SetPostStatus(slug, "invalid"); err == nil {
		t.Error("expected error for invalid status")
	}
	if err := st.SetPostStatus("no-such-slug", "draft"); err != sql.ErrNoRows {
		t.Errorf("err = %v", err)
	}
}

func TestAllTags_onlyPublished(t *testing.T) {
	st := openTestStore(t)
	tag := "alltags-" + strings.ToLower(RandomKeySuffix(6))
	pub := "alltags-pub-" + RandomKeySuffix(6)
	draft := "alltags-draft-" + RandomKeySuffix(6)
	now := time.Now().UTC()
	for slug, status := range map[string]string{pub: "published", draft: "draft"} {
		if err := st.UpsertPost(Post{
			Slug: slug, Title: "T", ContentMD: "m", ContentHTML: "h",
			Tags: []string{tag}, Status: status, PublishedAt: now, UpdatedAt: now,
		}); err != nil {
			t.Fatal(err)
		}
	}
	tags, err := st.AllTags()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, tg := range tags {
		if tg == tag {
			count++
		}
	}
	if count != 1 {
		t.Errorf("tag %q should appear once in AllTags, appearances=%d", tag, count)
	}
}

func TestCountPostsByStatus(t *testing.T) {
	st := openTestStore(t)
	totalBefore, _, _, err := st.CountPostsByStatus()
	if err != nil {
		t.Fatal(err)
	}
	slug := "count-" + RandomKeySuffix(8)
	now := time.Now().UTC()
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "C", ContentMD: "m", ContentHTML: "h",
		Status: "draft", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	total, pub, draft, err := st.CountPostsByStatus()
	if err != nil {
		t.Fatal(err)
	}
	if total != totalBefore+1 {
		t.Errorf("total = %d, want %d", total, totalBefore+1)
	}
	if draft < 1 {
		t.Errorf("draft count = %d", draft)
	}
	_ = pub
}
