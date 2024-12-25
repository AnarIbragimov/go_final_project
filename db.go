package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func initializeDB() (string, error) {
	dbName, ok := os.LookupEnv("TODO_DBFILE")
	if !ok {
		dbName = "scheduler.db"
	}

	appPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), dbName)

	var install bool
	_, err = os.Stat(dbFile)
	if err != nil {
		install = true
	}

	if !install {
		return dbName, nil
	}

	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return "", fmt.Errorf("failed to create or open db file: %w", err)
	}
	defer db.Close()

	query := `CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    comment TEXT,
    repeat VARCHAR(128)
	);
	CREATE INDEX idx_date ON scheduler(date);`

	if _, err = db.Exec(query); err != nil {
		return "", fmt.Errorf("failed to create table or index: %w", err)
	}

	return dbName, nil
}

func AddTask(dbName string, task Task) (int64, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return 0, fmt.Errorf("failed to open db file: %w", err)
	}
	defer db.Close()

	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to add a task into db: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetTasks(dbName string, search string) ([]Task, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return []Task{}, fmt.Errorf("failed to open db file: %w", err)
	}
	defer db.Close()

	var tasks []Task
	var query string
	var arguments []any

	if search == "" {
		query = "SELECT * FROM scheduler ORDER BY date LIMIT 50"
	} else if date, err := time.Parse("02.01.2006", search); err == nil {
		search = date.Format("20060102")
		query = "SELECT * FROM scheduler WHERE date = ? LIMIT 50"
		arguments = append(arguments, search)
	} else {
		query = "SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT 50"
		search = "%" + search + "%"
		arguments = append(arguments, search, search)
	}

	rows, err := db.Query(query, arguments...)
	if err != nil {
		return []Task{}, fmt.Errorf("failed to find tasks in DB: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return []Task{}, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return []Task{}, err
	}

	return tasks, nil
}

func GetTaskByID(dbName string, id string) (Task, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return Task{}, fmt.Errorf("failed to open db file: %w", err)
	}
	defer db.Close()

	var task Task

	row := db.QueryRow("SELECT * FROM scheduler WHERE id = ?", id)
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func UpdateTask(dbName string, task Task) error {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return fmt.Errorf("failed to open db file: %w", err)
	}
	defer db.Close()

	result, err := db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	if res, _ := result.RowsAffected(); res == 0 {
		return fmt.Errorf("No task found")
	}

	return nil
}

func DeleteTask(dbName string, id string) error {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return fmt.Errorf("failed to open db file: %w", err)
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}
	if res, _ := result.RowsAffected(); res == 0 {
		return fmt.Errorf("No task found")
	}

	return nil
}
