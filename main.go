// Пакет main - точка входа планировщика задач.
package main

import (
	"go_final_project/pkg/api"
	"go_final_project/pkg/db"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := 7540

	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err != nil {
			log.Fatalf("Incorrect port: %v", err)
		}
		port = p
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Error initializing DB: %v", err)
	}
	defer database.Close()

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", api.HandleNextDate)
	http.HandleFunc("/api/task", api.HandleTask(database))
	http.HandleFunc("/api/tasks", api.HandleTasks(database))
	http.HandleFunc("/api/task/done", api.HandleTaskDone(database))

	log.Printf("Running server on port %d", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
