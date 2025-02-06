package main

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/telegram"

	"fmt"

	_ "github.com/lib/pq" // Инициализация драйвера PostgreSQL
)

func main() {
	// Подключаемся к базе данных
	db, err := database.ConnectToDatabase()
	if err != nil {
		fmt.Println("Ошибка при подключении к базе данных:", err)
		return
	}
	defer db.Close() // Отложенное закрытие соединения с базой данных.

	fmt.Println("Успешное подключение к базе данных!")

	telegram.StartBot(db) // Запускаем Telegram бота.

}
