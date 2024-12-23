package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (t *Task) Validate() error {
	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
	if t.Title == "" {
		return fmt.Errorf("No title")
	}

	if t.Date == "" {
		t.Date = today.Format("20060102")
	}

	referenceDate, err := time.Parse("20060102", t.Date)
	if err != nil {
		return fmt.Errorf("Wrong date format: %s", t.Date)
	}

	if referenceDate.Before(today) && t.Repeat == "" {
		t.Date = time.Now().Format("20060102")
		return nil
	}

	t.Date, err = NextDate(today, t.Date, t.Repeat)
	if err != nil {
		return err
	}

	return nil
}

type App struct {
	WebDir string
	DB     string
}

func (app *App) TaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	switch r.Method {
	case http.MethodPost:
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		if err := task.Validate(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		id, err := AddTask(app.DB, task)
		if err != nil {
			errText := fmt.Sprintf(`{"error": "DB Error adding task: %s"}`, err.Error())
			http.Error(w, errText, http.StatusInternalServerError)
			return
		}

		response := map[string]int64{"id": id}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

	case http.MethodGet:
		tasks, _ := GetTasks(app.DB)
		fmt.Println(tasks)
		response := map[string][]Task{"tasks": tasks}
		if tasks == nil {
			response["tasks"] = []Task{}
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	}
}
