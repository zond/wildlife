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
	textTemplate "text/template"
	"html/template"
	"cells"
)

const (
	interval = time.Second * 2
)

func cellKey(cell *cells.Cell, r *http.Request) *datastore.Key {
	return datastore.NewKey(appengine.NewContext(r), "Cell", cell.Id(), 0, nil)
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
	http.HandleFunc("/js", js)
	http.HandleFunc("/css", css)
	http.HandleFunc("/load", load)
	http.HandleFunc("/click", click)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getXY(r *http.Request) (x, y []int) {
	err := r.ParseForm()
	if err != nil {
		panic(fmt.Errorf("While trying to parse form: %v", err))
	}
	for i := 0; i < min(len(r.Form["x"]), len(r.Form["y"])); i++ {
		j, err := strconv.Atoi(r.Form["x"][i])
		if err != nil {
			panic(fmt.Errorf("While trying to parse %s to int: %v", r.Form["x"][i], err))
		}
		j = j % cells.Width
		x = append(x, j)
		j, err = strconv.Atoi(r.Form["y"][i])
		if err != nil {
			panic(fmt.Errorf("While trying to parse %s to int: %v", r.Form["y"][i], err))
		}
		j = j % cells.Height
		y = append(y, j)
	}
	return
}

func click(w http.ResponseWriter, r *http.Request) {
	x, y := getXY(r)
	board := getCells(r)
	player := player(w, r)
	for i := 0; i < len(x); i++ {
		if cell, ok := board.Get(x[i], y[i]); ok {
			if cell.Player == player {
				removeCell(r, cell)
				delete(board, cell.Id())
			}
		} else {
			cell := &cells.Cell{x[i], y[i], player}
			putCell(r, cell)
			board[cell.Id()] = cell
		}
	}
	render(w, board)
}

func putCell(r *http.Request, cell *cells.Cell) {
	context := appengine.NewContext(r)
	if _, err := datastore.Put(context, cellKey(cell, r), cell); err != nil {
		panic(fmt.Errorf("While trying to store %v: %v", cell, err))
	}
}

func removeCell(r *http.Request, cell *cells.Cell) {
	context := appengine.NewContext(r)
	if err := datastore.Delete(context, cellKey(cell, r)); err != nil {
		panic(fmt.Errorf("While trying to delete %v: %v", cellKey(cell, r), err))
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

func tick(r *http.Request, board cells.CellMap, meta *Meta) cells.CellMap {
	meta.LastTick = time.Now()
	storeMeta(r, meta)
	rval := board.Tick()
	for _, cell := range rval {
		if !board.Has(cell) {
			putCell(r, cell)
		}
	}
	for _, cell := range board {
		if !rval.Has(cell) {
			removeCell(r, cell)
		}
	}
	return rval
} 

func getCells(r *http.Request) cells.CellMap {
	rval := make(cells.CellMap)
	context := appengine.NewContext(r)
	query := datastore.NewQuery("Cell")
	iterator := query.Run(context)
	for {
		cell := &cells.Cell{}
		if _, err := iterator.Next(cell); err == nil {
			rval[cell.Id()] = cell
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

func render(w http.ResponseWriter, board cells.CellMap) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(board.ToJson()); err != nil {
		panic(fmt.Errorf("While trying to encode %v: %v", board, err))
	}
}

func load(w http.ResponseWriter, r *http.Request) {
	render(w, getCells(r))
}

var htmlTemplates = template.Must(template.New("html").ParseGlob("templates/*.html"))
var jsTemplates = textTemplate.Must(textTemplate.New("js").ParseGlob("templates/*.js"))
var cssTemplates = textTemplate.Must(textTemplate.New("js").ParseGlob("templates/*.css"))

var sessionStore = sessions.NewCookieStore([]byte("wildlife in africa, we've got lions"))

func js(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	data := struct {
		Width int
		Height int
		Delay int
	}{cells.Width, cells.Height, int(interval / time.Millisecond)}
	if err := jsTemplates.ExecuteTemplate(w, "index.js", data); err != nil {
		panic(fmt.Errorf("While rendering index.js: %v", err))
	}
}

func css(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	if err := cssTemplates.ExecuteTemplate(w, "index.css", nil); err != nil {
		panic(fmt.Errorf("While rendering index.css: %v", err))
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	cols := make([]interface{}, cells.Width)
	rows := make([]interface{}, cells.Height)
	data := struct {
		Cols []interface{}
		Rows []interface{}
	}{cols, rows}
	if err := htmlTemplates.ExecuteTemplate(w, "index.html", data); err != nil {
		panic(fmt.Errorf("While rendering index.html: %v", err))
	}
}
