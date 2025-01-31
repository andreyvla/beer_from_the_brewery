package database

import (
	"beer_from_the_brewery/models"
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectToDatabase() (*sql.DB, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке .env файла: %w", err)
	}
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
	rows, err := db.Query("SELECT id, name, price, quantity, type, image_url, description FROM beers")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запросsа: %w", err)
	}
	defer rows.Close()

	var beers []models.Beer
	for rows.Next() {
		var beer models.Beer
		err := rows.Scan(&beer.ID, &beer.Name, &beer.Price, &beer.Quantity, &beer.Type, &beer.Description, &beer.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении данных: %w", err)
		}
		beers = append(beers, beer)
	}

	return beers, nil
}
