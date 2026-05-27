package server

import "time"

// Test hooks for white-box unit tests in package server_test.
func ParseDateForTest(s string) time.Time  { return parseDate(s) }
func TotalPagesForTest(total, limit int) int { return totalPages(total, limit) }
func TrimSpaceForTest(s string) string     { return trimSpace(s) }
