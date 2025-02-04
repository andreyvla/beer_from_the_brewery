package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/utils"
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleSearchCallback(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Введите название пива для поиска:")
	bot.Send(msg)
	waitingForSearchQuery[message.Chat.ID] = true
}

func handleSearchMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	searchQuery := message.Text
	foundBeers, err := database.SearchBeers(db, searchQuery)
	if err != nil {
		sendMessage(bot, message.Chat.ID, "Ошибка при поиске пива.", "", nil)
		return
	}

	if len(foundBeers) == 0 {
		sendMessage(bot, message.Chat.ID, "Пиво не найдено.", "", nil)
		return
	} else if len(foundBeers) == 1 {
		beer := foundBeers[0]
		msgText := utils.FormatBeerInfo(beer, true)

		// Создаем клавиатуру с одной кнопкой "Добавить в корзину"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Добавить в корзину %s", beer.Name), fmt.Sprintf("add_to_cart:%d:1", beer.ID)),
			),
		)
		sendMessage(bot, message.Chat.ID, msgText, "Markdown", &keyboard)

	} else { // Найдено несколько пив
		var beerListText string
		var beerRows [][]tgbotapi.InlineKeyboardButton
		for _, beer := range foundBeers {
			beerInfo := utils.FormatBeerInfo(beer, false)
			beerListText += fmt.Sprintf("%s\n\n", beerInfo)

			// Добавляем кнопку "Добавить в корзину" для каждого найденного пива
			beerRows = append(beerRows, []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Добавить в корзину %s", beer.Name), fmt.Sprintf("add_to_cart:%d:1", beer.ID)),
			})
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(beerRows...)
		sendMessage(bot, message.Chat.ID, beerListText, "Markdown", &keyboard)
	}
}
