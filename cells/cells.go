
package cells

import (
	"github.com/zond/tools"
	"fmt"
)

const (
	Width = 64
	Height = 56
)

func cellId(x, y int) string {
	return fmt.Sprintf("%v,%v", x, y)
}

type CellMap map[string]*Cell
func (self CellMap) Get(x, y int) (cell *Cell, ok bool) {
	cell, ok = self[cellId(x, y)]
	return
}
func (self CellMap) eachNeighbour(x, y int, f CellFunc) {
	for dx := -1; dx < 2; dx++ {
		for dy := -1; dy < 2; dy++ {
			if !(dx == 0 && dy == 0) {
				otherX := x + dx
				otherY := y + dy
				if otherX > 0 && otherX < Width && otherY > 0 && otherY < Height {
					if cell, ok := self[cellId(otherX, otherY)]; ok {
						f(cell)
					}
				}
			}
		}
	}
}
func (self CellMap) Has(cell *Cell) bool {
	if other, ok := self[cell.Id()]; ok {
		return other.Player == cell.Player
	}
	return false
}
func (self CellMap) countNeighbours(cell *Cell, onlyFriendly bool) int {
	rval := 0
	self.eachNeighbour(cell.X, cell.Y, func(otherCell *Cell) {
		if !onlyFriendly || cell.Player == otherCell.Player {
			rval++
		}
	})
	return rval
}
func (self CellMap) Tick() CellMap {
	rval := make(CellMap)
	for x := 0; x < Width; x++ {
		for y := 0; y < Height; y++ {
			if cell, ok := self.Get(x, y); ok {
				good := self.countNeighbours(cell, true)
				bad := self.countNeighbours(cell, false)
				if good > 1 && bad < 4 {
					rval[cell.Id()] = cell
				}
			} else {
				neigh := self.neighbourPlayers(x, y)
				if aspirants, ok := neigh[3]; ok && len(aspirants) == 1 {
					newCell := &Cell{x, y, aspirants[0]}
					rval[newCell.Id()] = newCell
				}
			}
		}
	}
	return rval
}
func (self CellMap) ToJson() CellMap {
	rval := make(CellMap)
	for key, val := range self {
		rval[key] = val.ToJson()
	}
	return rval
}
func (self CellMap) neighbourPlayers(x, y int) map[int][]string {
	counter := make(map[string]int)
	self.eachNeighbour(x, y, func(cell *Cell) {
		if count, ok := counter[cell.Player]; ok {
			counter[cell.Player] = count + 1
		} else {
			counter[cell.Player] = 1
		}
	})
	rval := make(map[int][]string)
	for player, count := range counter {
		if current, ok := rval[count]; ok {
			rval[count] = append(current, player)
		} else {
			rval[count] = []string{player}
		}
	}
	return rval
}

type CellFunc func(cell *Cell)

type Cell struct {
	X int
	Y int
	Player string
}
func (self *Cell) Id() string {
	return cellId(self.X, self.Y)
}
func (self *Cell) ToJson() *Cell {
	runes := []rune(self.Player)
	r := tools.NewBigIntInt(int(runes[len(runes) - 1])).BaseString(16)
	if len(r) < 2 {
		r = fmt.Sprint("0", r)
	}
	b := tools.NewBigIntInt(int(runes[len(runes) - 2])).BaseString(16)
	if len(b) < 2 {
		b = fmt.Sprint("0", b)
	}
	g := tools.NewBigIntInt(int(runes[len(runes) - 3])).BaseString(16)
	if len(g) < 2 {
		g = fmt.Sprint("0", g)
	}
	return &Cell{self.X, self.Y, fmt.Sprint("#", r, b, g)}
}