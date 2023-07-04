package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kraserh/energozvit/internal/tui"

	"github.com/kraserh/energozvit/internal/storage"
)

var Version string

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 2 && len(os.Args) != 4 {
		usageAndExit()
	}

	// Перший параметр: імʼя бази даних
	pathDB := os.Args[1]
	dir, file := filepath.Split(pathDB)
	if dir != "" {
		err := os.Chdir(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Другий параметр необовʼязковий: --create, для створення БД
	if len(os.Args) == 4 {
		if os.Args[2] != "--create" {
			usageAndExit()
		}
		const dateFormat = "2006-01-02"
		date := fmt.Sprintf("%s-01", os.Args[3])
		initDate, err := time.Parse(dateFormat, date)
		if err != nil {
			log.Fatal("Bad date format, expect YYYY-MM")
		}
		err = storage.Create(file, initDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Відкриття БД і запуск інтерфейса
	if len(os.Args) == 2 {
		stor, err := storage.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer stor.Close()

		tui.Start(stor)
	}
}

func usageAndExit() {
	fmt.Println("EnergoZvit programm")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Usage:\n  energozvit db_file [--create YYYY-MM]\n")
	os.Exit(0)
}
