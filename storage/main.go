package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
)

// DB представляє базу даних.
type DB struct {
	*sql.DB
}

// CreateDB створює базу даних з вказаним іменем і початковою датою.
// При помилкі завершує програму.
func CreateDB(filePath string, date Date) *DB {
	//
	if filePath == "" {
		err := errors.New("не вказано імʼя файла бази даних")
		check(err)
	}

	// Перевірка існування файла бази даних.
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		err := fmt.Errorf("файл %s вже існує", filePath)
		check(err)
	}

	// Створення БД
	var err error
	db := new(DB)
	db.DB, err = sql.Open("sqlite3", filePath)
	check(err)

	// Створення таблиць
	_, err = db.Exec(string(dbSchema))
	check(err)

	// Встановлення початкової дати
	query := `
		UPDATE stat
		SET value = ?
		WHERE key = ?`
	_, err = db.Exec(query, date.timestring(), "mDate")
	check(err)
	_, err = db.Exec(query, date.timestring(), "pDate")
	check(err)
	return db
}

// OpenDB відкриває базу даних з вказаним іменем.
// При помилкі завершує програму.
func OpenDB(filePath string) *DB {
	// Перевірка існування файла бази даних.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := fmt.Errorf("файл %s не існує", filePath)
		check(err)
	}

	// Створити резервну копію БД.
	cmd := exec.Command("cp", filePath, filePath+"~")
	err := cmd.Run()
	check(err)

	// Відкриття бази даних.
	db := new(DB)
	db.DB, err = sql.Open("sqlite3", filePath)
	check(err)

	// Перевірка версії бази даних
	var version string
	err = db.QueryRow("SELECT value FROM stat WHERE key=?",
		"version").Scan(&version)
	check(err)
	if version != "0.2" {
		err := errors.New("версія бази даних не підтримується")
		check(err)
	}
	return db
}

// Close закриває базу даних.
func (db *DB) Close() error {
	err := db.Close()
	return err
}

// check перериває програму якщо err містить помилку. Якщо вказано msg,
// то заміть err виводиться вказане повідомлення.
func check(err error, msg ...interface{}) {
	if err != nil {
		if len(msg) > 0 {
			fmt.Fprintln(os.Stderr, msg...)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(2)
	}
}
