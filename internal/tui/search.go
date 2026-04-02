package tui

import "strings"

// matchesQuery returns true if any field contains the query (case-insensitive).
// An empty query matches everything.
func matchesQuery(query string, fields ...string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}

// clampCursor ensures cursor stays within [0, length-1].
// Returns 0 for empty lists or negative cursors.
func clampCursor(cursor, length int) int {
	if length == 0 || cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

// handleSearchKey processes a key press during active search input.
// Returns the updated search string, whether search is still active, and whether the filter changed.
func handleSearchKey(key string, search string) (newSearch string, stillSearching bool, changed bool) {
	switch key {
	case "esc":
		return "", false, search != ""
	case "enter":
		return search, false, false
	case "backspace":
		if len(search) > 0 {
			return search[:len(search)-1], true, true
		}
		return search, true, false
	default:
		if len(key) == 1 {
			return search + key, true, true
		}
		return search, true, false
	}
}
