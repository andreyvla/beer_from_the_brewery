package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"sync"

	"database/sql"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var beers []models.Beer
var beersMutex = &sync.Mutex{}
var waitingForSearchQuery = make(map[int64]bool)
var carts sync.Map

func StartBot(db *sql.DB) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не задан!")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Авторизован как @%s", bot.Self.UserName)
	// Инициализируем список пива при запуске
	beersMutex.Lock()
	beers, err = database.GetBeers(db)
	if err != nil {
		log.Printf("Ошибка при начальной загрузке списка пива: %s", err.Error())
	}
	beersMutex.Unlock()
	go database.UpdateBeerList(db, &beers, beersMutex)

	updates := getUpdatesChannel(bot)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			handleCommand(bot, update.Message, db)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery, db)
		} else if update.Message != nil && !update.Message.IsCommand() {
			handleMessage(bot, update.Message, db)
		}
	}
}
func getUpdatesChannel(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Критическая ошибка: не удалось получить канал обновлений: %s", err)
	}
	return updates
}
