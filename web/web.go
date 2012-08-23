package web

import (
	"fmt"
	"net/http"
	"github.com/zond/tools"
	"code.google.com/p/gorilla/sessions"
	"encoding/json"
	"appengine"
	"appengine/datastore"
	"strconv"
	"time"
)

const (
	width = 128
	height = 96
	interval = time.Second
)

func cellId(x, y int) string {
	return fmt.Sprintf("%i,%i", x, y)
}

func cellKey(x, y int, r *http.Request) *datastore.Key {
	return datastore.NewKey(appengine.NewContext(r), "Cell", cellId(x, y), 0, nil)
}

type cellMap map[string]*Cell
func (self cellMap) eachNeighbour(x, y int, f cellFunc) {
	for dx := -1; dx < 1; dx++ {
		for dy := -1; dy < 1; dy++ {
			otherX := x + dx
			otherY := y + dy
			if otherX > 0 && otherX < width && otherY > 0 && otherY < height {
				if cell, ok := self[cellId(otherX, otherY)]; ok {
					f(cell)
				}
			}
		}
	}
}
func (self cellMap) countNeighbours(cell *Cell, onlyFriendly bool) int {
	rval := 0
	self.eachNeighbour(cell.X, cell.Y, func(otherCell *Cell) {
		if !onlyFriendly || cell.Player == otherCell.Player {
			rval++
		}
	})
	return rval
}
func (self cellMap) neighbourPlayers(x, y int) map[int][]string {
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

type cellFunc func(cell *Cell)

type Cell struct {
	X int
	Y int
	Player string
}
func (self *Cell) id() string {
	return cellId(self.X, self.Y)
}
func (self *Cell) key(r *http.Request) *datastore.Key {
	return cellKey(self.X, self.Y, r)
}

type Meta struct {
	LastTick time.Time
}

func getMetaKey(r *http.Request) *datastore.Key {
	context := appengine.NewContext(r)
	return datastore.NewKey(context, "Meta", "Meta", 0, nil)
}

func getMeta(r *http.Request) *Meta {
	context := appengine.NewContext(r)
	rval := &Meta{}
	if err := datastore.Get(context, getMetaKey(r), rval); err != nil && err != datastore.ErrNoSuchEntity {
		panic(fmt.Errorf("While trying to load meta: %v", err))
	}
	return rval
}

func storeMeta(r *http.Request, meta *Meta) {
	context := appengine.NewContext(r)
	if _, err := datastore.Put(context, getMetaKey(r), meta); err != nil {
		panic(fmt.Errorf("While trying to store %v: %v", meta, err))
	}
}

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/load", load)
	http.HandleFunc("/click", click)
}

func getXY(r *http.Request) (x, y int) {
	err := r.ParseForm()
	if err != nil {
		panic(fmt.Errorf("While trying to parse form: %v", err))
	}
	x, err = strconv.Atoi(r.Form["x"][0])
	if err != nil {
		panic(fmt.Errorf("While trying to parse %s to int: %v", r.Form["x"], err))
	}
	x = x % width
	y, err = strconv.Atoi(r.Form["y"][0])
	if err != nil {
		panic(fmt.Errorf("While trying to parse %s to int: %v", r.Form["y"], err))
	}
	y = y % height
	return
}

func click(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)
	x, y := getXY(r)
	key := cellKey(x, y, r)
	player := player(w, r)
	cell := &Cell{}
	if err := datastore.Get(context, key, cell); err != nil {
		if err != datastore.ErrNoSuchEntity {
			panic(fmt.Errorf("While trying to load cell %i,%i: %v", x, y, err))
		}
		cell := &Cell{x, y, player}
		putCell(r, cell)
	} else {
		if cell.Player == player {
			removeCell(r, cell)
		}
	}
	load(w, r)
}

func putCell(r *http.Request, cell *Cell) {
	context := appengine.NewContext(r)
	if _, err := datastore.Put(context, cell.key(r), cell); err != nil {
		panic(fmt.Errorf("While trying to store %v: %v", cell, err))
	}
}

func removeCell(r *http.Request, cell *Cell) {
	context := appengine.NewContext(r)
	if err := datastore.Delete(context, cell.key(r)); err != nil {
		panic(fmt.Errorf("While trying to delete %v: %v", cell.key(r), err))
	}
}

func player(w http.ResponseWriter, r *http.Request) string {
	session, err := sessionStore.Get(r, "wildlife")
	if err != nil {
		panic(fmt.Errorf("While trying to get session: %v", err))
	}
	if x, ok := session.Values["player"]; ok {
		return x.(string)
	}
	player := tools.Uuid()
	session.Values["player"] = player
	session.Save(r, w)
	return player
}

func tick(r *http.Request, cells cellMap, meta *Meta) cellMap {
	rval := make(cellMap)
	meta.LastTick = time.Now()
	storeMeta(r, meta)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if cell, ok := cells[cellId(x, y)]; ok {
				good := cells.countNeighbours(cell, true)
				bad := cells.countNeighbours(cell, false)
				if good < 2 {
					removeCell(r, cell)
				} else if bad > 3 {
					removeCell(r, cell)
				} else {
					rval[cell.id()] = cell
				}
			} else {
				neigh := cells.neighbourPlayers(x, y)
				if aspirants, ok := neigh[3]; ok && len(aspirants) == 1 {
					newCell := &Cell{x, y, aspirants[0]}
					rval[newCell.id()] = newCell
					putCell(r, newCell)
				}
			}
		}
	}
	return rval
} 

func getCells(r *http.Request) cellMap {
	rval := make(cellMap)
	context := appengine.NewContext(r)
	query := datastore.NewQuery("Cell")
	iterator := query.Run(context)
	cell := &Cell{}
	for {
		if _, err := iterator.Next(cell); err == nil {
			rval[cell.id()] = cell
		} else if err == datastore.Done {
			break
		} else {
			panic(fmt.Errorf("While trying to load next Cell: %v", err))
		}
	}
	meta := getMeta(r)
	if time.Now().Sub(meta.LastTick) > interval {
		rval = tick(r, rval, meta)
	}
	return rval
}

func load(w http.ResponseWriter, r *http.Request) {
	cells := getCells(r)
	rval := make([]*Cell, 0)
	for _, cell := range cells {
		rval = append(rval, cell)
	}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(rval); err != nil {
		panic(fmt.Errorf("While trying to encode %v: %v", rval, err))
	}
}

var sessionStore = sessions.NewCookieStore([]byte("wildlife in africa, we've got lions"))

func index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, fmt.Sprintf(`
<html>
<head>
<title>Wildlife</title>
</head>
<body>
<h1>Wildlife</h1>
</body>
</html>
			       `))
}
