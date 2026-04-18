package tui

import "testing"

func TestBulkSelection(t *testing.T) {
	t.Run("toggle selection", func(t *testing.T) {
		bs := newBulkState(5)
		bs.toggle(2)
		if !bs.isSelected(2) {
			t.Error("expected index 2 to be selected")
		}
		bs.toggle(2)
		if bs.isSelected(2) {
			t.Error("expected index 2 to be deselected")
		}
	})

	t.Run("select all", func(t *testing.T) {
		bs := newBulkState(3)
		bs.selectAll()
		if bs.count() != 3 {
			t.Errorf("expected 3 selected, got %d", bs.count())
		}
	})

	t.Run("deselect all when all selected", func(t *testing.T) {
		bs := newBulkState(3)
		bs.selectAll()
		bs.selectAll()
		if bs.count() != 0 {
			t.Errorf("expected 0 selected, got %d", bs.count())
		}
	})

	t.Run("selected indices", func(t *testing.T) {
		bs := newBulkState(5)
		bs.toggle(1)
		bs.toggle(3)
		indices := bs.selectedIndices()
		if len(indices) != 2 || indices[0] != 1 || indices[1] != 3 {
			t.Errorf("expected [1 3], got %v", indices)
		}
	})

	t.Run("reset clears all", func(t *testing.T) {
		bs := newBulkState(3)
		bs.selectAll()
		bs.reset()
		if bs.count() != 0 {
			t.Errorf("expected 0 after reset, got %d", bs.count())
		}
	})
}
