package main

import (
	"beer_from_the_brewery/database"
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	// Подключаемся к базе данных
	db, err := database.ConnectToDatabase()
	if err != nil {
		fmt.Println("Ошибка при вызове метода connectToDatabase: %w", err)
		return
	}
	defer db.Close()

	fmt.Println("Успешное подключение к базе данных!")

}
