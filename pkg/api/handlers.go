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

// taskRequest представляет тело запроса для создания и редактирования задачи.
type taskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// HandleNextDate обрабатывает запрос на вычисление следующей даты для повторения задачи.
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse(DateFormat, nowStr)
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
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func addTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req taskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("addTask: decode error: %v", err)
		jsonError(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format(DateFormat)

	if req.Date == "" || req.Date == "today" {
		req.Date = today
	}

	_, err := time.Parse(DateFormat, req.Date)
	if err != nil {
		log.Printf("addTask: time.Parse error: %v", err)
		jsonError(w, "invalid date", http.StatusBadRequest)
		return
	}

	if req.Repeat != "" {
		_, err := NextDate(now, req.Date, req.Repeat)
		if err != nil {
			log.Printf("addTask: NextDate error: %v", err)
			jsonError(w, "invalid repeat", http.StatusBadRequest)
			return
		}
	}

	if req.Date < today {
		if req.Repeat == "" {
			req.Date = today
		} else {
			next, err := NextDate(now, req.Date, req.Repeat)
			if err != nil {
				log.Printf("addTask: NextDate error: %v", err)
				jsonError(w, "invalid repeat", http.StatusBadRequest)
				return
			}
			req.Date = next
		}
	}

	id, err := appdb.AddTask(db, &appdb.Task{
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	})
	if err != nil {
		log.Printf("addTask: db error: %v", err)
		jsonError(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"id": strconv.FormatInt(id, 10)}); err != nil {
		log.Printf("addTask: encode error: %v", err)
	}
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{"error": msg}); err != nil {
		log.Printf("jsonError: encode error: %v", err)
	}
}

// HandleTasks возвращает список ближайших задач.
func HandleTasks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tasks, err := appdb.GetTasks(db)
		if err != nil {
			log.Printf("HandleTasks: GetTasks error: %v", err)
			jsonError(w, "db error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"tasks": tasks}); err != nil {
			log.Printf("HandleTasks: encode error: %v", err)
		}
	}
}

func getTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}
	task, err := appdb.GetTask(db, id)
	if err != nil {
		log.Printf("getTask: GetTask error: %v", err)
		jsonError(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("getTask: encode error: %v", err)
	}
}

func editTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req taskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("editTask: decode error: %v", err)
		jsonError(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format(DateFormat)

	if req.Date == "" || req.Date == "today" {
		req.Date = today
	}

	_, err := time.Parse(DateFormat, req.Date)
	if err != nil {
		log.Printf("editTask: time.Parse error: %v", err)
		jsonError(w, "invalid date", http.StatusBadRequest)
		return
	}

	if req.Repeat != "" {
		if _, err := NextDate(now, req.Date, req.Repeat); err != nil {
			log.Printf("editTask: NextDate error: %v", err)
			jsonError(w, "invalid repeat", http.StatusBadRequest)
			return
		}
	}

	if req.Date < today {
		if req.Repeat == "" {
			req.Date = today
		} else {
			next, err := NextDate(now, req.Date, req.Repeat)
			if err != nil {
				log.Printf("editTask: NextDate error: %v", err)
				jsonError(w, "invalid repeat", http.StatusBadRequest)
				return
			}
			req.Date = next
		}
	}

	if err := appdb.UpdateTask(db, &appdb.Task{
		ID:      req.ID,
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}); err != nil {
		log.Printf("editTask: UpdateTask error: %v", err)
		jsonError(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		log.Printf("editTask: encode error: %v", err)
	}
}

// HandleTaskDone отмечает задачу как выполненную.
func HandleTaskDone(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.FormValue("id")
		if id == "" {
			jsonError(w, "id is required", http.StatusBadRequest)
			return
		}

		task, err := appdb.GetTask(db, id)
		if err != nil {
			log.Printf("HandleTaskDone: GetTask error: %v", err)
			jsonError(w, "task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if task.Repeat == "" {
			if err := appdb.DeleteTask(db, id); err != nil {
				log.Printf("HandleTaskDone: DeleteTask error: %v", err)
				jsonError(w, "db error", http.StatusInternalServerError)
				return
			}
			if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
				log.Printf("HandleTaskDone: encode error: %v", err)
			}
			return
		}

		next, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			log.Printf("HandleTaskDone: NextDate error: %v", err)
			jsonError(w, "invalid repeat", http.StatusBadRequest)
			return
		}

		if err := appdb.UpdateTask(db, &appdb.Task{
			ID:      task.ID,
			Date:    next,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		}); err != nil {
			log.Printf("HandleTaskDone: UpdateTask error: %v", err)
			jsonError(w, "db error", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
			log.Printf("HandleTaskDone: encode error: %v", err)
		}
	}
}

func deleteTask(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		jsonError(w, "id is required", http.StatusBadRequest)
		return
	}
	if err := appdb.DeleteTask(db, id); err != nil {
		log.Printf("deleteTask: DeleteTask error: %v", err)
		jsonError(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		log.Printf("deleteTask: encode error: %v", err)
	}
}
