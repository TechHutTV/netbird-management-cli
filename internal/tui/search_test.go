package tui

import "testing"

func TestMatchesQuery(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		fields []string
		want   bool
	}{
		{"empty query matches everything", "", []string{"foo"}, true},
		{"exact match", "prod", []string{"production"}, true},
		{"case insensitive", "PROD", []string{"production"}, true},
		{"no match", "staging", []string{"production", "main"}, false},
		{"matches any field", "admin", []string{"web-server", "admin@co.com"}, true},
		{"partial match", "serv", []string{"web-server"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesQuery(tt.query, tt.fields...)
			if got != tt.want {
				t.Errorf("matchesQuery(%q, %v) = %v, want %v", tt.query, tt.fields, got, tt.want)
			}
		})
	}
}

func TestClampCursor(t *testing.T) {
	tests := []struct {
		name   string
		cursor int
		length int
		want   int
	}{
		{"within range", 3, 10, 3},
		{"at end", 9, 10, 9},
		{"past end", 15, 10, 9},
		{"empty list", 5, 0, 0},
		{"negative", -1, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampCursor(tt.cursor, tt.length)
			if got != tt.want {
				t.Errorf("clampCursor(%d, %d) = %d, want %d", tt.cursor, tt.length, got, tt.want)
			}
		})
	}
}

func TestHandleSearchKey(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		search          string
		wantSearch      string
		wantStill       bool
		wantChanged     bool
	}{
		{"esc clears search", "esc", "hello", "", false, true},
		{"esc on empty", "esc", "", "", false, false},
		{"enter keeps search", "enter", "hello", "hello", false, false},
		{"backspace removes char", "backspace", "hello", "hell", true, true},
		{"backspace on empty", "backspace", "", "", true, false},
		{"typing appends", "a", "hell", "hella", true, true},
		{"multi-char key ignored", "tab", "hello", "hello", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSearch, gotStill, gotChanged := handleSearchKey(tt.key, tt.search)
			if gotSearch != tt.wantSearch {
				t.Errorf("search = %q, want %q", gotSearch, tt.wantSearch)
			}
			if gotStill != tt.wantStill {
				t.Errorf("stillSearching = %v, want %v", gotStill, tt.wantStill)
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("changed = %v, want %v", gotChanged, tt.wantChanged)
			}
		})
	}
}
