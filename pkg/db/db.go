// Пакет db отвечает за инициализацию и работу с базой данных SQLite.
package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

// InitDB открывает базу данных и создаёт таблицу scheduler, если она не существует.
// Путь к файлу БД берётся из переменной окружения TODO_DBFILE или используется ./scheduler.db.
func InitDB() (*sql.DB, error) {
	dbFile := "./scheduler.db"
	if envFile := os.Getenv("TODO_DBFILE"); envFile != "" {
		dbFile = envFile
	}

	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	if install {
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS scheduler (
			id      INTEGER PRIMARY KEY AUTOINCREMENT,
			date    TEXT NOT NULL,
			title   TEXT NOT NULL,
			comment TEXT DEFAULT '',
			repeat  TEXT DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date)`)
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}
