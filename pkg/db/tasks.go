// Пакет db отвечает за инициализацию и работу с базой данных SQLite.
package db

import (
	"database/sql"
	"errors"
	"strconv"
)

const TasksLimit = 50

// AddTask добавляет новую задачу в базу данных и возвращает её id.
func AddTask(db *sql.DB, task *Task) (int64, error) {
	res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTasks возвращает список ближайших задач, отсортированных по дате.
func GetTasks(db *sql.DB) ([]*Task, error) {
	rows, err := db.Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*Task{}
	for rows.Next() {
		var id int64
		var task Task
		if err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTask возвращает задачу по id.
func GetTask(db *sql.DB, id string) (*Task, error) {
	var tid int64
	var task Task
	err := db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?`, id).
		Scan(&tid, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}
	task.ID = strconv.FormatInt(tid, 10)
	return &task, nil
}

// UpdateTask обновляет задачу по id. Возвращает ошибку, если задача не найдена.
func UpdateTask(db *sql.DB, task *Task) error {
	res, err := db.Exec(`UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`,
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
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
