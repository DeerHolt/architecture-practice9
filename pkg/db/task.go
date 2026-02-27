// Пакет db отвечает за инициализацию и работу с базой данных SQLite.
package db

// Task представляет задачу в планировщике.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
