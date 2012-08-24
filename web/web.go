package web

import (
	"fmt"
	"net/http"
	"github.com/zond/tools"
	"code.google.com/p/gorilla/sessions"
	"encoding/json"
	"strconv"
	"time"
	textTemplate "text/template"
	"html/template"
	"cells"
)

const (
	interval = time.Second * 2
	metaKey = "meta"
	cellsKey = "cells"
)

type Meta struct {
	LastTick time.Time
}

type cellMapContainer struct {
	cells cells.CellMap
}

var meta = &Meta{}
var board = make(cells.CellMap)

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
				delete(board, cell.Id())
			}
		} else {
			cell := &cells.Cell{x[i], y[i], player}
			board[cell.Id()] = cell
		}
	}
	render(w, board)
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

func getCells(r *http.Request) cells.CellMap {
	if time.Now().Sub(meta.LastTick) > interval {
		meta.LastTick = time.Now()
		board = board.Tick()
	}
	return board
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
