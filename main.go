package main

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/telegram"
	"log"
	"os"

	_ "github.com/lib/pq" // Инициализация драйвера PostgreSQL
)

func main() {
	// Создаем логгер, который пишет в stderr
	logger := log.New(os.Stderr, "beer_bot: ", log.LstdFlags|log.Lshortfile)

	// Подключаемся к базе данных
	db, err := database.ConnectToDatabase()
	if err != nil {
		logger.Fatalf("Ошибка при подключении к базе данных: %v", err) // Fatalf завершает программу
	}
	defer db.Close() // Отложенное закрытие соединения с базой данных.

	logger.Println("Успешное подключение к базе данных!")

	telegram.StartBot(db, logger) // Передаем логгер в StartBot
}
