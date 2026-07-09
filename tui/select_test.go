package tui

import (
	"strings"
	"testing"

	"mop/scanner"
)

// newTestModel builds a Model with the given groups, initialising the derived
// entries/selection state the way scanResultMsg does.
func newTestModel(groups []scanner.ToolGroup) Model {
	m := Model{
		groups:    groups,
		selected:  make(map[[2]int]bool),
		whitelist: make(map[string]bool),
		height:    24,
		width:     80,
	}
	m.buildEntries()
	for gi, g := range groups {
		for ii := range g.Items {
			m.selected[[2]int{gi, ii}] = true
		}
	}
	return m
}

// TestEntryLinesMatchRenderedLines locks the single-source-of-truth invariant:
// entryLines[i] returned by buildContentLines must point at the actual rendered
// line for entry i. This is the class of bug that hid WorkBuddy at the bottom of
// the list: an empty group above shifts every following line, and the old
// cursorLine() re-derivation forgot the "(No caches found)" row, so scrolling
// drifted and the bottom group became unreachable.
func TestEntryLinesMatchRenderedLines(t *testing.T) {
	groups := []scanner.ToolGroup{
		{Name: "EmptyTool", Type: "CLI", Items: nil}, // empty group above
		{Name: "BigTool", Type: "CLI", Items: []scanner.CacheItem{
			{Path: "/x/a", Size: 300, Description: "Traces"},
			{Path: "/x/b", Size: 100, Description: "Logs"},
		}},
	}
	m := newTestModel(groups)

	lines, entryLines := buildContentLines(m)

	if len(entryLines) != len(m.entries) {
		t.Fatalf("entryLines len %d != entries len %d", len(entryLines), len(m.entries))
	}

	// Every entry's recorded line must actually render that entry: headers show
	// the group name, items show the item description.
	for i, e := range m.entries {
		ln := entryLines[i]
		if ln < 0 || ln >= len(lines) {
			t.Fatalf("entry %d line %d out of range [0,%d)", i, ln, len(lines))
		}
		want := groups[e.groupIdx].Name
		if !e.isHeader {
			want = groups[e.groupIdx].Items[e.itemIdx].Description
		}
		if !strings.Contains(lines[ln], want) {
			t.Errorf("entry %d expected line %d to contain %q, got %q", i, ln, want, lines[ln])
		}
	}
}

// TestBottomGroupReachableAfterEmptyGroups verifies the concrete WorkBuddy
// symptom: with empty groups above, the last item's rendered line must be within
// the content so scrolling can reveal it (maxOffset covers it).
func TestBottomGroupReachableAfterEmptyGroups(t *testing.T) {
	groups := []scanner.ToolGroup{
		{Name: "EmptyA", Type: "CLI", Items: nil},
		{Name: "EmptyB", Type: "CLI", Items: nil},
		{Name: "EmptyC", Type: "CLI", Items: nil},
		{Name: "WorkBuddy", Type: "CLI + Desktop", Items: []scanner.CacheItem{
			{Path: "/w/traces", Size: 284, Description: "Traces"},
			{Path: "/w/logs", Size: 96, Description: "Logs"},
		}},
	}
	m := newTestModel(groups)

	lines, entryLines := buildContentLines(m)
	lastEntry := len(m.entries) - 1
	lastLine := entryLines[lastEntry]

	// The last entry must map to a real rendered line containing its content.
	if lastLine >= len(lines) || !strings.Contains(lines[lastLine], "Logs") {
		t.Fatalf("last entry line %d does not render the bottom item (lines=%d)", lastLine, len(lines))
	}

	// Simulate moving the cursor to the last entry and scrolling: the offset must
	// be able to bring the last line into view.
	m.cursor = lastEntry
	m.updateScrollOffset()
	visible := calcVisibleHeight(m)
	if lastLine < m.scrollOffset || lastLine >= m.scrollOffset+visible {
		t.Errorf("bottom item line %d not visible with offset %d, height %d", lastLine, m.scrollOffset, visible)
	}
}
