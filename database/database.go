package database

import (
	"beer_from_the_brewery/models"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL
)

// ConnectToDatabase устанавливает соединение с базой данных.
func ConnectToDatabase() (*sql.DB, error) {
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке .env файла: %w", err)
	}
	// Получение параметров подключения из переменных окружения
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

// GetBeers получает список всего пива из базы данных.
func GetBeers(db *sql.DB) ([]models.Beer, error) {
	rows, err := db.Query("SELECT id, name, price, quantity, type, image_url, description FROM beers")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запросsа: %w", err)
	}
	defer rows.Close()

	var beers []models.Beer
	for rows.Next() {
		var beer models.Beer
		err := rows.Scan(&beer.ID, &beer.Name, &beer.Price, &beer.Quantity, &beer.Type, &beer.ImageURL, &beer.Description)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении данных: %w", err)
		}
		beers = append(beers, beer)
	}

	return beers, nil
}

// SearchBeers ищет пиво по названию в базе данных.
func SearchBeers(db *sql.DB, searchQuery string) ([]models.Beer, error) {
	rows, err := db.Query("SELECT id, name, price, quantity, type, image_url, description FROM beers WHERE lower(name) LIKE lower($1)", "%"+searchQuery+"%")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	var beers []models.Beer
	for rows.Next() {
		var beer models.Beer
		err := rows.Scan(&beer.ID, &beer.Name, &beer.Price, &beer.Quantity, &beer.Type, &beer.ImageURL, &beer.Description)
		if err != nil {
			return nil, fmt.Errorf("ошибка при чтении данных: %w", err)
		}
		beers = append(beers, beer)
	}

	return beers, nil
}

// GetBeerByID получает информацию о пиве по его ID.
func GetBeerByID(db *sql.DB, beerID int) (*models.Beer, error) {
	var beer models.Beer
	err := db.QueryRow("SELECT id, name, price, quantity, type, image_url, description FROM beers WHERE id = $1", beerID).Scan(&beer.ID, &beer.Name, &beer.Price, &beer.Quantity, &beer.Type, &beer.ImageURL, &beer.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении пива по ID: %w", err)
	}
	return &beer, nil

}

// UpdateBeerList периодически обновляет список доступного пива.
func UpdateBeerList(db *sql.DB, beers *[]models.Beer, beersMutex *sync.Mutex) {
	for {
		newBeers, err := GetBeers(db)
		if err != nil {
			log.Printf("Ошибка при обновлении списка пива: %s", err.Error())
		} else {
			beersMutex.Lock()
			*beers = newBeers
			beersMutex.Unlock()
		}

		time.Sleep(5 * time.Minute)
	}
}

// CreateOrder создает новый заказ в базе данных.
func CreateOrder(db *sql.DB, userID int64, cartItems []models.CartItem) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	// Создаем запись в таблице orders, используя RETURNING id
	orderDate := time.Now()
	orderStatus := "new"

	var orderID int64 // Объявляем переменную для хранения orderID
	row := tx.QueryRow("INSERT INTO orders (user_id, order_date, status) VALUES ($1, $2, $3) RETURNING id", userID, orderDate, orderStatus)
	if err := row.Scan(&orderID); err != nil { // Считываем orderID из результата запроса
		return fmt.Errorf("не удалось получить ID заказа: %w", err)
	}

	// Создаем записи в таблице order_items для каждого товара в корзине
	for _, cartItem := range cartItems {
		_, err = tx.Exec("INSERT INTO order_items (order_id, beer_id, quantity) VALUES ($1, $2, $3)", orderID, cartItem.BeerID, cartItem.Quantity)
		if err != nil {
			return fmt.Errorf("не удалось добавить позицию заказа: %w", err)
		}
	}

	return tx.Commit() // Фиксируем транзакцию, если всё прошло успешно
}
