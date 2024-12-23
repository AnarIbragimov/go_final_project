package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

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

func GetTasks(dbName string) ([]Task, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return []Task{}, fmt.Errorf("failed to open db file: %w", err)
	}
	defer db.Close()

	var tasks []Task

	query := "SELECT * FROM scheduler ORDER BY date LIMIT 50"
	rows, err := db.Query(query)
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
