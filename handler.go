package main

import (
	"database/sql"
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
		t.Date = today.Format(format)
	}

	referenceDate, err := time.Parse(format, t.Date)
	if err != nil {
		return fmt.Errorf("Wrong date format: %s", t.Date)
	}

	if referenceDate.Before(today) {
		t.Date = today.Format(format)
	}

	if t.Repeat == "" {
		return nil
	}

	if _, err := NextDate(today, t.Date, t.Repeat); err != nil {
		return err
	}

	return nil
}

type App struct {
	WebDir   string
	DB       *sql.DB
	Password string
}

func (app *App) TaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	id := r.FormValue("id")

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
		if id == "" {
			http.Error(w, `{"error": "Не указан идентификатор"}`, http.StatusBadRequest)
			return
		}

		task, err := GetTaskByID(app.DB, id)
		if err != nil {
			http.Error(w, `{"error": "Задача не найдена"}`, http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(task); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

	case http.MethodPut:
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		if err := task.Validate(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		err = UpdateTask(app.DB, task)
		if err != nil {
			http.Error(w, `{"error": "Task not found"}`, http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		err := DeleteTask(app.DB, id)
		if err != nil {
			http.Error(w, `{"error": "Could not delete task"}`, http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) TasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	search := r.FormValue("search")
	tasks, _ := GetTasks(app.DB, search)

	response := map[string][]Task{"tasks": tasks}
	if tasks == nil {
		response["tasks"] = []Task{}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
}

func (app *App) TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	id := r.FormValue("id")

	task, err := GetTaskByID(app.DB, id)
	if err != nil {
		http.Error(w, `{"error": "Задача не найдена"}`, http.StatusBadRequest)
		return
	}

	if err := task.Validate(); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	switch task.Repeat {
	case "":
		err = DeleteTask(app.DB, id)
		if err != nil {
			http.Error(w, `{"error": "Could not delete task"}`, http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	default:
		now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		task.Date, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error": "Could not update task"}`, http.StatusInternalServerError)
			return
		}
		err = UpdateTask(app.DB, task)
		if err != nil {
			http.Error(w, `{"error": "Could not update task"}`, http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) TaskNextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nowTime, err := time.Parse(format, now)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	result, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
	}

	w.Write([]byte(result))
}

func (app *App) SignInHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error": Server error}`, http.StatusInternalServerError)
		return
	}

	if v, ok := m["password"]; !ok {
		http.Error(w, `{"error": Server error}`, http.StatusInternalServerError)
		return
	} else if v != app.Password {
		http.Error(w, `{"error": Wrong password}`, http.StatusUnauthorized)
		return
	}

	token, err := createJWTToken(app.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]any{"token": token}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": Server error}`, http.StatusInternalServerError)
		return
	}
}
