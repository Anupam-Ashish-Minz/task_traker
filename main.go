package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	HoursRemaining float32
}

func getTasks() ([]Task, error) {
	db, err := sql.Open(DB_TYPE, DB_NAME)
	defer db.Close()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`select id, name, time_started, hours_alloted, hours_completed from tasks`)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		rows.Scan(&task.ID, &task.Name, &task.TimeStarted, &task.HoursAlloted, &task.HoursCompleted)
		task.HoursRemaining = task.HoursAlloted - task.HoursCompleted
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/taskbody.html")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks, err := getTasks()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, tasks)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("invalid add task request")
		return
	}

	log.Println("add task request")

	name := r.PostFormValue("name")
	hours_alloted := r.PostFormValue("hours_alloted")
	time_started := time.Now().Format(time.RFC3339)
	hours_completed := 0

	if name == "" || hours_alloted == "" || hours_alloted == "0" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := sql.Open(DB_TYPE, DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := db.Exec(`insert into tasks (name, time_started, hours_alloted, hours_completed) values (?, ?, ?, ?)`,
		name, time_started, hours_alloted, hours_completed)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	taskID, _ := res.LastInsertId()

	row := db.QueryRow(`select id, name, time_started, hours_alloted, hours_completed from tasks where id = ?`, taskID)
	var task Task
	row.Scan(&task.ID, &task.Name, &task.TimeStarted, &task.HoursAlloted, &task.HoursCompleted)

	tmpl, err := template.ParseFiles("templates/taskbody.html")
	tmpl.Execute(w, task)
}

func completeTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("should be a post method"))
		return
	}
	p := strings.Split(r.URL.Path, "/")
	taskID, err := strconv.Atoi(p[len(p)-1])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("task id required"))
		return
	}

	db, err := sql.Open(DB_TYPE, DB_NAME)
	defer db.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to connect to database"))
		return
	}

	_, err = db.Exec(`update tasks set hours_completed = hours_completed + ? where id = ?`, 1, taskID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to execute query"))
		return
	}

	row := db.QueryRow(`select id, name, time_started, hours_alloted, hours_completed from tasks where id = ?`, taskID)
	var task Task
	row.Scan(&task.ID, &task.Name, &task.TimeStarted, &task.HoursAlloted, &task.HoursCompleted)
	task.HoursRemaining = task.HoursAlloted - task.HoursCompleted

	tmpl, err := template.ParseFiles("templates/taskbody.html")
	tmpl.Execute(w, task)
}

func main() {
	staticDir := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	)

	http.Handle("/static/", staticDir)
	http.HandleFunc("/add", addTask)
	http.HandleFunc("/task/complete_one_hour/", completeTask)
	http.HandleFunc("/", index)

	http.ListenAndServe(":4000", nil)
}
