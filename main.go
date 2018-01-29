package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kraserh/energozvit/storage"
)

var (
	Version   string
	BuildTime string
)

func main() {
	switch len(os.Args) {
	case 1:
		storageCreate()
	case 2:
		storageOpen(os.Args[1])
	default:
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [path_to_database]\n",
			os.Args[0])
	}
}

func storageCreate() {
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
	fmt.Println(date, filePath)
}

func storageOpen(filePath string) {
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
