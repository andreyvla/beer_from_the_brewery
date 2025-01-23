package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"sync"

	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var beers []models.Beer
var beersMutex = &sync.Mutex{}
var waitingForSearchQuery = make(map[int64]bool)

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
					tgbotapi.NewInlineKeyboardButtonData("Показать пиво", "beer"),
					tgbotapi.NewInlineKeyboardButtonData("Найти пиво", "search"),
				),
			)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот для покупки пива. Выберите действие:")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		} else if update.CallbackQuery != nil { //  Обработка нажатия на кнопку
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.AnswerCallbackQuery(callback); err != nil {
				log.Printf("Ошибка при ответе на callback: %s", err.Error())
			}

			switch update.CallbackQuery.Data {
			case "beer":

				if len(beers) == 0 {
					sendMessage(bot, update.CallbackQuery.Message.Chat.ID, "Пива нет :(", "")
					continue
				}
				beerList := formatBeerList(beers, "")
				sendMessage(bot, update.CallbackQuery.Message.Chat.ID, beerList, "Markdown")
			case "search": // Обработчик кнопки "Найти пиво"
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введите название пива для поиска:")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true) // Убираем клавиатуру после нажатия на кнопку поиска
				bot.Send(msg)

				// Устанавливаем состояние ожидания поискового запроса
				waitingForSearchQuery[update.CallbackQuery.Message.Chat.ID] = true
			}

		} else if update.Message != nil && !update.Message.IsCommand() {
			// Проверяем, ожидает ли бот поисковый запрос от пользователя
			if waitingForSearchQuery[update.Message.Chat.ID] {
				searchQuery := update.Message.Text
				beerList := formatBeerList(beers, searchQuery)
				sendMessage(bot, update.Message.Chat.ID, beerList, "Markdown")
				delete(waitingForSearchQuery, update.Message.Chat.ID) // Сбрасываем состояние ожидания
			}

		}
	}
}

func formatBeerList(beers []models.Beer, searchQuery string) string {
	beersMutex.Lock()
	defer beersMutex.Unlock()

	var beerList string

	for _, beer := range beers {
		if searchQuery == "" || containsIgnoreCase(beer.Name, searchQuery) || containsIgnoreCase(beer.Type, searchQuery) {
			beerList += fmt.Sprintf("*%d. %s - %s*\nЦена: %.2f\nВ наличии: %d\n\n", beer.ID, beer.Name, beer.Type, beer.Price, beer.Quantity)
		}
	}

	if beerList == "" && searchQuery != "" {
		return "Пиво не найдено."
	}

	return beerList
}
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
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

		time.Sleep(5 * time.Minute)
	}
}
