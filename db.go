package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

func initializeDB() (string, error) {
	dbName, ok := os.LookupEnv("TODO_DBFILE")
	if !ok {
		dbName = "scheduler.db"
	}

	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return "", fmt.Errorf("failed to create or open db file: %w", err)
	}
	defer db.Close()

	query := `CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    comment TEXT,
    repeat VARCHAR(128)
	);
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);`

	if _, err = db.Exec(query); err != nil {
		return "", fmt.Errorf("failed to create table or index: %w", err)
	}

	return dbName, nil
}

func AddTask(db *sql.DB, task Task) (int64, error) {
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

func GetTasks(db *sql.DB, search string) ([]Task, error) {
	var tasks []Task
	var query string
	var arguments []any

	if search == "" {
		query = "SELECT (id, date, title, comment, repeat) FROM scheduler ORDER BY date LIMIT ?"
		arguments = append(arguments, limit)
	} else if date, err := time.Parse("02.01.2006", search); err == nil {
		search = date.Format(format)
		query = "SELECT * FROM scheduler WHERE date = ? LIMIT ?"
		arguments = append(arguments, search, limit)
	} else {
		query = "SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?"
		search = "%" + search + "%"
		arguments = append(arguments, search, search, limit)
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

func GetTaskByID(db *sql.DB, id string) (Task, error) {
	var task Task

	row := db.QueryRow("SELECT (id, date, title, comment, repeat) FROM scheduler WHERE id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func UpdateTask(db *sql.DB, task Task) error {
	result, err := db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	if res, _ := result.RowsAffected(); res == 0 {
		return fmt.Errorf("No task found")
	}

	return nil
}

func DeleteTask(db *sql.DB, id string) error {
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}
	if res, _ := result.RowsAffected(); res == 0 {
		return fmt.Errorf("No task found")
	}

	return nil
}
