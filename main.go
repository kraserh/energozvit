package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kraserh/energozvit/gui"
	"github.com/kraserh/energozvit/storage"
)

var (
	// Version - версія програми
	Version string
	// BuildTime - дата збірки програми
	BuildTime string
)

func main() {
	switch len(os.Args) {
	case 1:
		filePath := storageCreate()
		storageOpen(filePath)
	case 2:
		storageOpen(os.Args[1])
	default:
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [path_to_database]\n",
			os.Args[0])
	}
}

func storageCreate() string {
	stdin := bufio.NewReader(os.Stdin)
	// filePath
	pwd, _ := os.Getwd()
	fmt.Printf("Поточний каталог: %s\n", pwd)
	fmt.Print("Введіть шлях до створюємого файла бази даних: ")
	filePath, err := stdin.ReadString('\n')
	check(err)
	filePath = filePath[:len(filePath)-1]
	// date
	var date storage.Date
	for {
		fmt.Print("Введіть початкову дату (YYYY-MM): ")
		dateStr, _ := stdin.ReadString('\n')
		if date, err = storage.ParseDate(dateStr); err == nil {
			break
		}
		fmt.Println("Невірний формат дати")
	}
	storage.CreateDB(filePath, date)
	return filePath
}

func storageOpen(filePath string) {
	db := storage.OpenDB(filePath)
	defer db.Close()
	gui.Start(db)
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
		os.Exit(1)
	}
}
