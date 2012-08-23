
package cells

import (
	"testing"
)

func TestCountNeighbours(t *testing.T) {
	cells := make(CellMap)
	cell := &Cell{10, 10, "1"}
	cells[cell.Id()] = cell
	if count := cells.countNeighbours(cell, true); count != 0 {
		t.Errorf("%v should have 0 neigbours in %v but had %v", cell, cells, count)
	}
	if count := cells.countNeighbours(cell, false); count != 0 {
		t.Errorf("%v should have 0 neigbours in %v but had %v", cell, cells, count)
	}
	cell = &Cell{10, 11, "1"}
	cells[cell.Id()] = cell
	if count := cells.countNeighbours(cell, true); count != 1 {
		t.Errorf("%v should have 0 neigbours in %v but had %v", cell, cells, count)
	}
	if count := cells.countNeighbours(cell, false); count != 1 {
		t.Errorf("%v should have 0 neigbours in %v but had %v", cell, cells, count)
	}
	cell = &Cell{11, 10, "2"}
	cells[cell.Id()] = cell
	if count := cells.countNeighbours(cell, true); count != 0 {
		t.Errorf("%v should have 0 neigbours in %v but had %v", cell, cells, count)
	}
	if count := cells.countNeighbours(cell, false); count != 2 {
		t.Errorf("%v should have 0 neigbours in %v but had %v", cell, cells, count)
	}
}


func TestTick(t *testing.T) {
	cells := make(CellMap)
	cell1 := &Cell{10, 10, "1"}
	cells[cell1.Id()] = cell1
	cells2 := cells.Tick()
	if len(cells2) != 0 {
		t.Errorf("%v should not have any cells", cells2)
	}
	cell2 := &Cell{10, 11, "1"}
	cells[cell2.Id()] = cell2
	cell3 := &Cell{11, 10, "1"}
	cells[cell3.Id()] = cell3
	cell4 := &Cell{11, 11, "1"}
	cells[cell4.Id()] = cell4
	cells2 = cells.Tick()
	if len(cells2) != 4 {
		t.Errorf("%v should have 4 cells", cells2)
	}
	if _, ok := cells2[cell1.Id()]; !ok {
		t.Errorf("%v should have %v", cells2, cell1)
	}
	if _, ok := cells2[cell2.Id()]; !ok {
		t.Errorf("%v should have %v", cells2, cell2)
	}
	if _, ok := cells2[cell3.Id()]; !ok {
		t.Errorf("%v should have %v", cells2, cell3)
	}
	if _, ok := cells2[cell4.Id()]; !ok {
		t.Errorf("%v should have %v", cells2, cell4)
	}
}

func TestBlinker(t *testing.T) {
	cells := make(CellMap)
	cell1 := &Cell{10, 9, "1"}
	cells[cell1.Id()] = cell1
	cell2 := &Cell{10, 10, "1"}
	cells[cell2.Id()] = cell2
	cell3 := &Cell{10, 11, "1"}
	cells[cell3.Id()] = cell3
	cells2 := cells.Tick()
	if len(cells2) != 3 {
		t.Errorf("%v should have 3 cells", cells2)
	}
	if _, ok := cells2[cellId(9, 10)]; !ok {
		t.Errorf("%v should have a cell at 9,10")
	}
	if _, ok := cells2[cellId(10, 10)]; !ok {
		t.Errorf("%v should have a cell at 10,10")
	}
	if _, ok := cells2[cellId(11, 10)]; !ok {
		t.Errorf("%v should have a cell at 11,10")
	}
}