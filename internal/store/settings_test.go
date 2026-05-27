package store

import (
	"testing"
	"time"
)

func TestSettings_roundTrip(t *testing.T) {
	st := openTestStore(t)
	if err := st.SetSetting("site_title", "My Site"); err != nil {
		t.Fatal(err)
	}
	v, err := st.GetSetting("site_title")
	if err != nil || v != "My Site" {
		t.Fatalf("GetSetting: %q err=%v", v, err)
	}
	if err := st.UpdateSettings(map[string]string{
		"site_title":     "Updated",
		"posts_per_page": "7",
	}); err != nil {
		t.Fatal(err)
	}
	all, err := st.GetAllSettings()
	if err != nil {
		t.Fatal(err)
	}
	if all["site_title"] != "Updated" || all["posts_per_page"] != "7" {
		t.Errorf("settings = %v", all)
	}
}

func TestApplySiteConfig_overwritesSettings(t *testing.T) {
	st := openTestStore(t)
	_ = st.SetSetting("site_title", "Old Title")
	if err := st.ApplySiteConfig("From Config", "desc", "https://example.com", 12); err != nil {
		t.Fatal(err)
	}
	title, err := st.GetSetting("site_title")
	if err != nil || title != "From Config" {
		t.Fatalf("site_title = %q err=%v", title, err)
	}
	if n := st.PostsPerPageFromSettings(10); n != 12 {
		t.Errorf("posts_per_page = %d", n)
	}
}

func TestPostsPerPageFromSettings(t *testing.T) {
	st := openTestStore(t)
	if n := st.PostsPerPageFromSettings(10); n != 10 {
		t.Errorf("default = %d", n)
	}
	_ = st.SetSetting("posts_per_page", "15")
	if n := st.PostsPerPageFromSettings(10); n != 15 {
		t.Errorf("from settings = %d", n)
	}
	_ = st.SetSetting("posts_per_page", "bad")
	if n := st.PostsPerPageFromSettings(10); n != 10 {
		t.Errorf("invalid = %d", n)
	}
}

func TestLastPostUpdate(t *testing.T) {
	st := openTestStore(t)
	slug := "last-update-" + RandomKeySuffix(8)
	now := time.Date(2026, 3, 15, 8, 0, 0, 0, time.UTC)
	if err := st.UpsertPost(Post{
		Slug: slug, Title: "L", ContentMD: "m", ContentHTML: "h",
		Status: "published", PublishedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	ts, err := st.LastPostUpdate()
	if err != nil {
		t.Fatal(err)
	}
	if ts == nil {
		t.Fatal("expected last update time")
	}
}
