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
)

type Cell struct {
	X int
	Y int
	Player string
}
func (self *Cell) Id() string {
	return fmt.Sprintf("%i,%i", self.X, self.Y)
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
	if err := datastore.Get(context, getMetaKey(r), rval); err != nil {
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
	http.HandleFunc("/tick", tick)
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
	key := datastore.NewKey(context, "Cell", (&Cell{x, y, ""}).Id(), 0, nil)
	player := player(w, r)
	cell := &Cell{}
	if err := datastore.Get(context, key, cell); err != nil {
		if err != datastore.ErrNoSuchEntity {
			panic(fmt.Errorf("While trying to load cell %i,%i: %v", x, y, err))
		}
		cell := &Cell{x, y, player}
		if _, err = datastore.Put(context, key, cell); err != nil {
			panic(fmt.Errorf("While trying to store %v: %v", cell, err))
		}
	} else {
		if cell.Player == player {
			if err = datastore.Delete(context, key); err != nil {
				panic(fmt.Errorf("While trying to delete %v: %v", key, err))
			}
		}
	}
	tick(w, r)
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

func getCells(r *http.Request) map[string]*Cell {
	rval := make(map[string]*Cell)
	context := appengine.NewContext(r)
	query := datastore.NewQuery("Cell")
	iterator := query.Run(context)
	cell := &Cell{}
	for {
		if _, err := iterator.Next(cell); err == nil {
			rval[cell.Id()] = cell
		} else if err == datastore.Done {
			break
		} else {
			panic(fmt.Errorf("While trying to load next Cell: %v", err))
		}
	}
	
	return rval
}

func tick(w http.ResponseWriter, r *http.Request) {
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
