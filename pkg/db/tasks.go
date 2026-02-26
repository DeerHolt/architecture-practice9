// Пакет db отвечает за инициализацию и работу с базой данных SQLite.
package db

import (
	"database/sql"
	"errors"
	"strconv"
)

// AddTask добавляет новую задачу в базу данных и возвращает её id.
func AddTask(db *sql.DB, date, title, comment, repeat string) (int64, error) {
	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		date, title, comment, repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTasks возвращает список ближайших задач, отсортированных по дате.
func GetTasks(db *sql.DB) ([]map[string]string, error) {
	rows, err := db.Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []map[string]string{}
	for rows.Next() {
		var id int64
		var date, title, comment, repeat string
		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, map[string]string{
			"id":      strconv.FormatInt(id, 10),
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		})
	}
	return tasks, nil
}

// GetTask возвращает задачу по id.
func GetTask(db *sql.DB, id string) (map[string]string, error) {
	var tid int64
	var date, title, comment, repeat string
	err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?`, id).
		Scan(&tid, &date, &title, &comment, &repeat)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"id":      strconv.FormatInt(tid, 10),
		"date":    date,
		"title":   title,
		"comment": comment,
		"repeat":  repeat,
	}, nil
}

// UpdateTask обновляет задачу по id. Возвращает ошибку, если задача не найдена.
func UpdateTask(db *sql.DB, id, date, title, comment, repeat string) error {
	res, err := db.Exec(`UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`,
		date, title, comment, repeat, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("task not found")
	}
	return nil
}

// DeleteTask удаляет задачу по id. Возвращает ошибку, если задача не найдена.
func DeleteTask(db *sql.DB, id string) error {
	res, err := db.Exec(`DELETE FROM scheduler WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("task not found")
	}
	return nil
}
