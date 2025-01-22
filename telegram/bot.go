package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"sync"

	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var beers []models.Beer //  Глобальная переменная для хранения списка пива
var beersMutex = &sync.Mutex{}

func StartBot(db *sql.DB) {
	// Получаем токен бота из переменных окружения
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не задан!")
	}

	// Создаем новый бот
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
		// Здесь можно добавить обработку ошибки, например, завершить работу бота
	}
	beersMutex.Unlock()
	go updateBeerList(db)

	// Конфигурируем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	// Обрабатываем входящие обновления
	for update := range updates {
		// Обрабатываем команду /start
		if update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start" {
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Показать пиво", "beer"), //  Data - это то, что вернется в callback
				),
			)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот для покупки пива. Выберите действие:")
			msg.ReplyMarkup = keyboard //  Прикрепляем клавиатуру к сообщению
			bot.Send(msg)
		} else if update.CallbackQuery != nil { //  Обработка нажатия на кнопку
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.AnswerCallbackQuery(callback); err != nil {
				log.Printf("Ошибка при ответе на callback: %s", err.Error())
			}

			switch update.CallbackQuery.Data {
			case "beer":

				if len(beers) == 0 {
					sendMessage(bot, update.CallbackQuery.Message.Chat.ID, "Пива нет :(", "") // parseMode пустая для обычного текста
					continue
				}
				beerList := formatBeerList(beers)
				sendMessage(bot, update.CallbackQuery.Message.Chat.ID, beerList, "Markdown") // parseMode "Markdown"

			}
		}
	}
}

func formatBeerList(beers []models.Beer) string {
	beersMutex.Lock()
	defer beersMutex.Unlock()
	var beerList string
	for _, beer := range beers {
		beerList += fmt.Sprintf("*%s*\n%s\nЦена: %d\nВ наличии: %d\n\n", beer.Name, beer.Description, beer.Price, beer.Quantity)
	}
	return beerList
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string, parseMode string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %s", err.Error())
	}
}

func updateBeerList(db *sql.DB) {
	for {
		beersMutex.Lock() // Блокируем доступ к beers
		newBeers, err := database.GetBeers(db)
		if err != nil {
			log.Printf("Ошибка при обновлении списка пива: %s", err.Error())
		} else {
			beers = newBeers // Обновляем список пива
		}
		beersMutex.Unlock() // Разблокируем доступ к beers

		time.Sleep(5 * time.Minute) // Обновляем список каждые 5 минут
	}
}
