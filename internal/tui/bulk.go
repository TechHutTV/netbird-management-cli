package tui

import "sort"

type bulkState struct {
	selected map[int]bool
	total    int
}

func newBulkState(total int) *bulkState {
	return &bulkState{
		selected: make(map[int]bool),
		total:    total,
	}
}

func (b *bulkState) toggle(index int) {
	if b.selected[index] {
		delete(b.selected, index)
	} else {
		b.selected[index] = true
	}
}

func (b *bulkState) isSelected(index int) bool {
	return b.selected[index]
}

func (b *bulkState) selectAll() {
	if b.count() == b.total {
		b.selected = make(map[int]bool)
	} else {
		newSelected := make(map[int]bool, b.total)
		for i := 0; i < b.total; i++ {
			newSelected[i] = true
		}
		b.selected = newSelected
	}
}

func (b *bulkState) count() int {
	return len(b.selected)
}

func (b *bulkState) selectedIndices() []int {
	indices := make([]int, 0, len(b.selected))
	for i := range b.selected {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	return indices
}

func (b *bulkState) reset() {
	b.selected = make(map[int]bool)
}
