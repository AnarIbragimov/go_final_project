package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	dbName, err := initializeDB()
	if err != nil {
		log.Fatal(err)
	}

	app := &App{DB: dbName, WebDir: "./web"}
	http.Handle("/", http.FileServer(http.Dir(app.WebDir)))
	http.HandleFunc("/api/nextdate", app.TaskNextDateHandler)

	port, ok := os.LookupEnv("TODO_PORT")
	if !ok {
		port = "7540"
	}

	password := os.Getenv("TODO_PASSWORD")
	if password != "" {
		http.HandleFunc("/api/signin", app.SignInHandler)
		http.HandleFunc("/api/tasks", auth(app.TasksHandler))
		http.HandleFunc("/api/task", auth(app.TaskHandler))
		http.HandleFunc("/api/task/done", auth(app.TaskDoneHandler))
	} else {
		http.HandleFunc("/api/tasks", app.TasksHandler)
		http.HandleFunc("/api/task", app.TaskHandler)
		http.HandleFunc("/api/task/done", app.TaskDoneHandler)
	}

	log.Printf("Serving files on http://localhost:%s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
