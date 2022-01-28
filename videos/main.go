package main

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

const limit = 30

type api struct {
	db *sqlx.DB
}

type Video struct {
	Id    string `db:"id"`
	Title string `db:"title"`
	Views int64  `db:"views"`
}

func (a *api) videos(w http.ResponseWriter, r *http.Request) {
	var videos []Video

	err := a.db.SelectContext(r.Context(), &videos, `SELECT id, title, views FROM videos ORDER BY views LIMIT ?`, limit)
	if err != nil {
		a.fail(w, "failed to fetch posts: "+err.Error(), 500)
		return
	}

	data := struct {
		Videos []Video
	}{videos}

	a.ok(w, data)
}

func main() {
	dsn := "user=postgres password=postgres dbname=my_database sslmode=disable"
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	app := &api{db: db}
	http.HandleFunc("/posts", app.videos)
	http.ListenAndServe(":8080", nil)
}

func (a *api) fail(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")

	data := struct {
		Error string
	}{Error: msg}

	resp, _ := json.Marshal(data)
	w.WriteHeader(status)
	w.Write(resp)
}

func (a *api) ok(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.fail(w, "oops something evil has happened", 500)
		return
	}
	w.Write(resp)
}
