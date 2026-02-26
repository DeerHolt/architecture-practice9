// Пакет api содержит HTTP-обработчики и вспомогательные функции для работы с API.
package api

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// DateFormat — формат даты, используемый в планировщике.
const DateFormat = "20060102"

// NextDate вычисляет следующую дату для задачи по правилу повторения.
// Поддерживаемые правила: y — ежегодно, d N — каждые N дней (1-400).
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat is empty")
	}

	start, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(repeat)

	switch parts[0] {
	case "y":
		next := start.AddDate(1, 0, 0)
		for !next.After(now) {
			next = next.AddDate(1, 0, 0)
		}
		return next.Format(DateFormat), nil

	case "d":
		if len(parts) < 2 {
			return "", errors.New("d requires number")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n < 1 || n > 400 {
			return "", errors.New("invalid d value")
		}
		next := start.AddDate(0, 0, n)
		for !next.After(now) {
			next = next.AddDate(0, 0, n)
		}
		return next.Format(DateFormat), nil

	default:
		return "", errors.New("unknown repeat rule")
	}
}
