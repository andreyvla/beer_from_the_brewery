package database

import (
	"beer_from_the_brewery/models"
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
		return nil, fmt.Errorf("ошибка при загрузке .env файла: %w", err)
	}
	// Извлекаем значения из переменных окружения
	user, ok := os.LookupEnv("POSTGRES_USER")
	if !ok || user == "" {
		return nil, fmt.Errorf("переменная окружения POSTGRES_USER не задана")
	}
	pass, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok || pass == "" {
		return nil, fmt.Errorf("переменная окружения POSTGRES_PASSWORD не задана")
	}
	host, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok || host == "" {
		return nil, fmt.Errorf("переменная окружения POSTGRES_HOST не задана")
	}
	port, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok || port == "" {
		return nil, fmt.Errorf("переменная окружения POSTGRES_PORT не задана")
	}
	dbname, ok := os.LookupEnv("POSTGRES_DB")
	if !ok || dbname == "" {
		return nil, fmt.Errorf("переменная окружения POSTGRES_DB не задана")
	}

	// Формируем строку подключения, используя переменные окружения
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии соединения с базой данных: %w", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке соединения с базой данных: %w", err)
	}

	return db, nil
}
func GetBeers(db *sql.DB) ([]models.Beer, error) {
	rows, err := db.Query("SELECT id, name, description, price, quantity, image_url FROM beers")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	var beers []models.Beer
	for rows.Next() {
		var beer models.Beer
		err := rows.Scan(&beer.ID, &beer.Name, &beer.Description, &beer.Price, &beer.Quantity, &beer.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении данных: %w", err)
		}
		beers = append(beers, beer)
	}

	return beers, nil
}
