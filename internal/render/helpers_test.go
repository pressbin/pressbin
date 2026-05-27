package render

import "testing"

func TestStripHTML(t *testing.T) {
	in := "<p>Hello <strong>world</strong> &amp; friends</p>"
	got := stripHTML(in)
	if got != "Hello  world  & friends" {
		t.Errorf("stripHTML = %q", got)
	}
}

func TestTrimWords(t *testing.T) {
	short := "one two three"
	if got := trimWords(short, 5); got != short {
		t.Errorf("short text changed: %q", got)
	}
	long := "one two three four five six"
	got := trimWords(long, 3)
	if got != "one two three…" {
		t.Errorf("trimWords = %q", got)
	}
}
