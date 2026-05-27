package server_test

import (
	"testing"
	"time"

	"pressbin.dev/pressbin/internal/server"
)

func TestParseDate(t *testing.T) {
	rfc := "2026-01-15T10:30:00Z"
	got := server.ParseDateForTest(rfc)
	want, _ := time.Parse(time.RFC3339, rfc)
	if !got.Equal(want.UTC()) {
		t.Errorf("RFC3339: got %v want %v", got, want.UTC())
	}
	dateOnly := server.ParseDateForTest("2026-06-20")
	if dateOnly.Year() != 2026 || dateOnly.Month() != 6 || dateOnly.Day() != 20 {
		t.Errorf("date only: %v", dateOnly)
	}
	empty := server.ParseDateForTest("")
	if empty.IsZero() {
		t.Error("empty date should default to now")
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		total, limit, want int
	}{
		{0, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{25, 10, 3},
		{5, 0, 1},
	}
	for _, tc := range tests {
		if got := server.TotalPagesForTest(tc.total, tc.limit); got != tc.want {
			t.Errorf("totalPages(%d,%d) = %d, want %d", tc.total, tc.limit, got, tc.want)
		}
	}
}

func TestTrimSpace(t *testing.T) {
	if got := server.TrimSpaceForTest("  hello\t "); got != "hello" {
		t.Errorf("trimSpace = %q", got)
	}
}
