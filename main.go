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
	http.HandleFunc("/api/tasks", app.TaskHandler)
	http.HandleFunc("/api/task", app.TaskHandler)

	port, ok := os.LookupEnv("TODO_PORT")
	if !ok {
		port = "7540"
	}

	log.Printf("Serving files on http://localhost:%s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
