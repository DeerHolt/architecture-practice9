// Пакет api содержит HTTP-обработчики и вспомогательные функции для работы с API.
package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	appdb "go_final_project/pkg/db"
)

// HandleNextDate обрабатывает запрос на вычисление следующей даты для повторения задачи.
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "invalid now", http.StatusBadRequest)
		return
	}

	next, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Write([]byte(next))
}

// HandleTask обрабатывает CRUD-операции над одной задачей.
func HandleTask(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			addTask(db, w, r)
		case http.MethodGet:
			getTask(db, w, r)
		case http.MethodPut:
			editTask(db, w, r)
		case http.MethodDelete:
			deleteTask(db, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func addTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var task struct {
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("addTask: decode error: %v", err)
		jsonError(w, "invalid json")
		return
	}

	if task.Title == "" {
		jsonError(w, "title is required")
		return
	}

	now := time.Now()
	today := now.Format("20060102")

	if task.Date == "" || task.Date == "today" {
		task.Date = today
	}

	_, err := time.Parse("20060102", task.Date)
	if err != nil {
		log.Printf("addTask: time.Parse error: %v", err)
		jsonError(w, "invalid date")
		return
	}

	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Printf("addTask: NextDate error: %v", err)
			jsonError(w, "invalid repeat")
			return
		}
	}

	if task.Date < today {
		if task.Repeat == "" {
			task.Date = today
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				log.Printf("addTask: NextDate error: %v", err)
				jsonError(w, "invalid repeat")
				return
			}
			task.Date = next
		}
	}

	id, err := appdb.AddTask(db, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		log.Printf("addTask: db error: %v", err)
		jsonError(w, "db error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"id": strconv.FormatInt(id, 10)})
}

func jsonError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"error": msg})
}

// HandleTasks возвращает список ближайших задач.
func HandleTasks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := appdb.GetTasks(db)
		if err != nil {
			log.Printf("HandleTasks: GetTasks error: %v", err)
			jsonError(w, "db error")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"tasks": tasks})
	}
}

func getTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		jsonError(w, "id is required")
		return
	}
	task, err := appdb.GetTask(db, id)
	if err != nil {
		log.Printf("getTask: GetTask error: %v", err)
		jsonError(w, "task not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func editTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var task struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("editTask: json.NewDecoder error: %v", err)
		jsonError(w, "invalid json")
		return
	}

	if task.ID == "" {
		jsonError(w, "id is required")
		return
	}
	if task.Title == "" {
		jsonError(w, "title is required")
		return
	}

	now := time.Now()
	today := now.Format("20060102")

	if task.Date == "" || task.Date == "today" {
		task.Date = today
	}

	_, err := time.Parse("20060102", task.Date)
	if err != nil {
		log.Printf("editTask: time.Parse error: %v", err)
		jsonError(w, "invalid date")
		return
	}

	if task.Repeat != "" {
		if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
			log.Printf("editTask: NextDate error: %v", err)
			jsonError(w, "invalid repeat")
			return
		}
	}

	if task.Date < today {
		if task.Repeat == "" {
			task.Date = today
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				log.Printf("editTask: NextDate error: %v", err)
				jsonError(w, "invalid repeat")
				return
			}
			task.Date = next
		}
	}

	if err := appdb.UpdateTask(db, task.ID, task.Date, task.Title, task.Comment, task.Repeat); err != nil {
		log.Printf("editTask: UpdateTask error: %v", err)
		jsonError(w, "task not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{})
}

// HandleTaskDone отмечает задачу как выполненную.
func HandleTaskDone(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			jsonError(w, "id is required")
			return
		}

		task, err := appdb.GetTask(db, id)
		if err != nil {
			log.Printf("HandleTaskDone: GetTask error: %v", err)
			jsonError(w, "task not found")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if task["repeat"] == "" {
			if err := appdb.DeleteTask(db, id); err != nil {
				jsonError(w, "db error")
				return
			}
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}

		next, err := NextDate(time.Now(), task["date"], task["repeat"])
		if err != nil {
			log.Printf("HandleTaskDone: NextDate error: %v", err)
			jsonError(w, "invalid repeat")
			return
		}

		if err := appdb.UpdateTask(db, id, next, task["title"], task["comment"], task["repeat"]); err != nil {
			log.Printf("HandleTaskDone: UpdateTask error: %v", err)
			jsonError(w, "db error")
			return
		}

		json.NewEncoder(w).Encode(map[string]any{})
	}
}

func deleteTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		jsonError(w, "id is required")
		return
	}
	if err := appdb.DeleteTask(db, id); err != nil {
		log.Printf("deleteTask: DeleteTask error: %v", err)
		jsonError(w, "task not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{})
}
