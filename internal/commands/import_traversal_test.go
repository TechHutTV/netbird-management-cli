package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoinUnder(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"plain filename", "groups.yml", false},
		{"dotdot escape", "../../../etc/passwd", true},
		{"nested dotdot escape", "a/../../../../etc/passwd", true},
		{"subdir ok", "sub/file.yml", false},
		{"empty string resolves to base", "", false},
		// Note: a leading "/" is stripped by filepath.Join (on POSIX),
		// so "/etc/passwd" becomes "{tmp}/etc/passwd" — still under base.
		// That's safe behavior, not an escape.
		{"leading slash gets absorbed into base", "/etc/passwd", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := safeJoinUnder(tmp, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("safeJoinUnder(%q): err=%v wantErr=%v", tc.input, err, tc.wantErr)
			}
			if err != nil {
				return
			}
			absTmp, _ := filepath.Abs(tmp)
			if !strings.HasPrefix(got, absTmp) {
				t.Errorf("result %q is not under %q", got, absTmp)
			}
		})
	}
}

func TestSafeJoinUnder_SamePathAsBase(t *testing.T) {
	// safeJoinUnder(base, ".") should produce base itself.
	tmp := t.TempDir()
	got, err := safeJoinUnder(tmp, ".")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	absTmp, _ := filepath.Abs(tmp)
	if got != absTmp {
		t.Errorf("got %q, want %q", got, absTmp)
	}
}

func TestSafeJoinUnder_DotDotWithSymlinkFails(t *testing.T) {
	// Ensures repeated `..` segments trying to escape are caught regardless
	// of surrounding literal path components.
	tmp := t.TempDir()
	cases := []string{
		"../outside",
		"sub/../../outside",
		"./..",
		"a/b/../../../outside",
	}
	for _, input := range cases {
		if _, err := safeJoinUnder(tmp, input); err == nil {
			t.Errorf("safeJoinUnder(%q) should have rejected an escape", input)
		}
	}
	_ = os.PathSeparator // ensure import is still used
}
