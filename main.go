package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_TYPE = "sqlite3"
	DB_NAME = "test.db"
)

type Task struct {
	ID             int
	Name           string
	TimeStarted    string
	HoursAlloted   float32
	HoursCompleted float32
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Println(err)
	}

	db, err := sql.Open(DB_TYPE, DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`select id, name, time_started, hours_alloted, hours_completed from tasks`)
	defer rows.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		rows.Scan(&task.ID, &task.Name, &task.TimeStarted, &task.HoursAlloted, &task.HoursCompleted)
		tasks = append(tasks, task)
	}

	tmpl.Execute(w, tasks)
}

func main() {
	http.HandleFunc("/", index)
	http.ListenAndServe(":4000", nil)
}
