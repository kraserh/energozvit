package main

import (
	"bufio"
	"fmt"
	"os"
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
	var filePath string
	for {
		pwd, _ := os.Getwd()
		fmt.Printf("Поточний каталог: %s\n", pwd)
		fmt.Print("Введіть шлях до створюємого файла бази даних:")
		if _, err := fmt.Scanln(&filePath); err == nil {
			break
		}
		stdin.ReadString('\n') // скидаєм буфер вводу
		fmt.Println("Помилка")
	}
	var dateStr string
	for {
		fmt.Print("Введіть початкову дату (YYYY-MM): ")
		if _, err := fmt.Scanln(&dateStr); err == nil {
			break
		}
		stdin.ReadString('\n')
		fmt.Println("Невірний формат дати")
	}
}

func storageOpen(filePath string) {
}
