package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"context"
	"sync"
	"time"

	"database/sql"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Глобальные переменные для хранения данных бота
var (
	beers                 []models.Beer          // Список доступного пива
	beersMutex            = &sync.Mutex{}        // Мьютекс для безопасного доступа к beers
	waitingForSearchQuery = make(map[int64]bool) // Карта для отслеживания пользователей, ожидающих результаты поиска
	carts                 sync.Map               // Карта для хранения корзин пользователей (ключ - chatID, значение - map[int]models.CartItem)
)

// StartBot запускает Telegram бота.
func StartBot(db *sql.DB, logger *log.Logger) {
	// Получаем токен бота из переменных окружения.
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не задан!")
	}

	// Создаем новый экземпляр бота.
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	logger.Printf("Авторизован как @%s", bot.Self.UserName)

	// Инициализируем список пива при запуске с контекстом и таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	beersMutex.Lock()
	beers, err = database.GetBeers(ctx, db)
	beersMutex.Unlock()

	if err != nil {
		logger.Printf("Ошибка при начальной загрузке списка пива: %s", err.Error())
	}

	// Запускаем горутину для периодического обновления списка пива с контекстом.
	go database.UpdateBeerList(context.Background(), db, &beers, beersMutex, logger) // Передаем контекст и логгер

	// Получаем канал обновлений от Telegram.
	updates := getUpdatesChannel(bot)

	// Обрабатываем обновления.
	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			handleCommand(bot, update.Message, db, logger)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery, db, logger)
		} else if update.Message != nil && !update.Message.IsCommand() {
			handleMessage(bot, update.Message, db, logger)
		}
	}
}

// getUpdatesChannel возвращает канал обновлений от Telegram.
func getUpdatesChannel(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Критическая ошибка: не удалось получить канал обновлений: %s", err)
	}
	return updates
}
