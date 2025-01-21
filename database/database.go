package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv" // Добавляем импорт
	_ "github.com/lib/pq"      // Импортируем драйвер PostgreSQL
)

func ConnectToDatabase() (*sql.DB, error) {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Ошибка при загрузке .env файла: %w", err)
	}
	// Извлекаем значения из переменных окружения
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return nil, fmt.Errorf("Переменная окружения POSTGRES_USER не задана")
	}
	pass := os.Getenv("POSTGRES_PASSWORD")
	if user == "" {
		return nil, fmt.Errorf("Переменная окружения POSTGRES_PASSWORD не задана")
	}
	host := os.Getenv("POSTGRES_HOST")
	if user == "" {
		return nil, fmt.Errorf("Переменная окружения POSTGRES_HOST не задана")
	}
	port := os.Getenv("POSTGRES_PORT")
	if user == "" {
		return nil, fmt.Errorf("Переменная окружения POSTGRES_PORT не задана")
	}
	dbname := os.Getenv("POSTGRES_DB")
	if user == "" {
		return nil, fmt.Errorf("Переменная окружения POSTGRES_DB не задана")
	}

	// Формируем строку подключения, используя переменные окружения
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при открытии соединения с базой данных: %w", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Ошибка при проверке соединения с базой данных: %w", err)
	}

	return db, nil
}
