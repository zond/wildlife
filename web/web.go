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
)

type Cell struct {
	X int
	Y int
	Player string
}

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/tick", tick)
	http.HandleFunc("/click", click)
}

func click(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)
	if err := r.ParseForm(); err != nil {
		panic(fmt.Errorf("While trying to parse form: %v", err))
	}
	x, err := strconv.Atoi(r.Form["x"][0])
	if err != nil {
		panic(fmt.Errorf("While trying to parse %s to int: %v", r.Form["x"], err))
	}
	y, err := strconv.Atoi(r.Form["y"][0])
	if err != nil {
		panic(fmt.Errorf("While trying to parse %s to int: %v", r.Form["y"], err))
	}
	key := datastore.NewKey(context, "Cell", fmt.Sprintf("%i,%i", x, y), 0, nil)
	player := player(w, r)
	cell := &Cell{}
	if err = datastore.Get(context, key, cell); err != nil {
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

func tick(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)
	query := datastore.NewQuery("Cell")
	iterator := query.Run(context)
	rval := make([]Cell, 0)
	cell := &Cell{}
	for {
		if _, err := iterator.Next(cell); err == nil {
			rval = append(rval, *cell)
		} else {
			encoder := json.NewEncoder(w)
			w.Header().Set("Content-Type", "application/json")
			if err = encoder.Encode(rval); err != nil {
				context.Errorf("While trying to return json encoded %s: %s", rval, err)
			}
			break
		}
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
