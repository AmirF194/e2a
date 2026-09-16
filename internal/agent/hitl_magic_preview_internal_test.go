package agent

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestTruncatePreviewNeverSplitsARune pins the byte cap at a boundary that
// falls mid-rune: a naive s[:magicPreviewMaxBytes] slice would split the
// rune, producing invalid UTF-8.
func TestTruncatePreviewNeverSplitsARune(t *testing.T) {
	// One byte short of the cap, then a 3-byte rune straddling the boundary.
	s := strings.Repeat("a", magicPreviewMaxBytes-1) + "界" + strings.Repeat("b", 100)

	got := truncatePreview(s)

	if !strings.HasSuffix(got, "truncated; view full in the dashboard.") {
		t.Fatalf("missing truncation notice, got tail %q", got[len(got)-60:])
	}
	kept := strings.TrimSuffix(got, "\n\n... truncated; view full in the dashboard.")
	if !utf8.ValidString(kept) {
		t.Fatalf("truncated preview is not valid UTF-8: %q", kept)
	}
	if len(kept) >= magicPreviewMaxBytes {
		t.Fatalf("kept %d bytes, want under the %d-byte cap", len(kept), magicPreviewMaxBytes)
	}
	if strings.Contains(kept, "界") {
		t.Fatalf("the straddling rune should have been cut, not kept whole past the boundary")
	}
}

// TestTruncatePreviewLeavesShortBodyUntouched is the control: under the cap,
// truncatePreview is a no-op.
func TestTruncatePreviewLeavesShortBodyUntouched(t *testing.T) {
	s := "plain body"
	if got := truncatePreview(s); got != s {
		t.Fatalf("truncatePreview(%q) = %q, want unchanged", s, got)
	}
}
