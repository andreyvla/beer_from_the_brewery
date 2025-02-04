package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"beer_from_the_brewery/utils"
	"fmt"
	"strconv"

	"database/sql"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	switch message.Command() {
	case "start":
		handleStartCommand(bot, message)
	default:
		sendMessage(bot, message.Chat.ID, "Неизвестная команда.", "", nil)
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	if strings.HasPrefix(callbackQuery.Data, "add_to_cart:") {
		handleAddToCartCallback(bot, callbackQuery, db)
	} else if strings.HasPrefix(callbackQuery.Data, "adjust_quantity:") {
		handleAdjustQuantityCallback(bot, callbackQuery, db)
	} else if strings.HasPrefix(callbackQuery.Data, "confirm_add:") {
		handleConfirmAddCallback(bot, callbackQuery, db)
	} else if callbackQuery.Data == "checkout" { // Обработка оформления заказа
		handleCheckoutCallback(bot, callbackQuery, db)
	} else if callbackQuery.Data == "clear_cart" { // Обработка очистки корзины
		handleClearCartCallback(bot, callbackQuery)
	} else {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неизвестное действие.", "", nil)
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот для покупки пива.")
	msg.ReplyMarkup = createMainKeyboard()
	bot.Send(msg)
}

func handleAddToCartCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	data := strings.Split(callbackQuery.Data, ":")
	if len(data) != 3 {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат данных.", "", nil)
		return
	}

	beerID, err := strconv.Atoi(data[1])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный ID пива.", "", nil)
		return
	}

	beer, err := database.GetBeerByID(db, beerID)
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil)
		return
	}

	if beer == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Пиво не найдено.", "", nil)
		return
	}

	keyboard := createQuantityKeyboard(beerID, 1) // Начальное количество 1

	sendMessage(bot, callbackQuery.Message.Chat.ID, fmt.Sprintf("Укажите количество %s:", beer.Name), "", &keyboard)

}

func handleAdjustQuantityCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {

	data := strings.Split(callbackQuery.Data, ":")
	if len(data) != 4 {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат данных.", "", nil)
		return
	}
	beerID, err := strconv.Atoi(data[1])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный ID пива.", "", nil)
		return
	}
	quantity, err := strconv.Atoi(data[2])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат количества.", "", nil)
		return
	}
	adjust, err := strconv.Atoi(data[3])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при изменении количества.", "", nil)
		return
	}

	newQuantity := quantity + adjust
	if newQuantity <= 0 {
		newQuantity = 1
	}
	keyboard := createQuantityKeyboard(beerID, newQuantity)
	editMsg := tgbotapi.NewEditMessageReplyMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, keyboard)
	_, err = bot.Send(editMsg)
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при обновлении сообщения", "", nil)
	}
}

func handleConfirmAddCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	data := strings.Split(callbackQuery.Data, ":")
	beerID, err := strconv.Atoi(data[1])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный ID пива.", "", nil)
		return
	}
	quantity, err := strconv.Atoi(data[2])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат количества.", "", nil)
		return
	}
	beer, err := database.GetBeerByID(db, beerID)
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil)
		return
	}
	if beer == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Пиво не найдено.", "", nil)
		return
	}
	// Получаем текущую корзину или создаем новую
	loadCart, _ := carts.LoadOrStore(callbackQuery.Message.Chat.ID, make(map[int]models.CartItem))
	cart := loadCart.(map[int]models.CartItem)

	// Обновляем количество или добавляем новую запись
	cartItem, ok := cart[beerID]
	if ok {
		cartItem.Quantity += quantity
	} else {
		cartItem = models.CartItem{BeerID: beerID, Quantity: quantity}
	}
	cart[beerID] = cartItem

	carts.Store(callbackQuery.Message.Chat.ID, cart) // Сохраняем обновленную корзину

	sendMessage(bot, callbackQuery.Message.Chat.ID, fmt.Sprintf("%s (%d шт.) добавлен в корзину.", beer.Name, quantity), "", nil)
}

func handleBeerCallback(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	beersMutex.Lock()   // Добавляем блокировку
	beersList := beers  // Создаем копию
	beersMutex.Unlock() // Освобождаем блокировку
	var beerListText string

	if len(beersList) == 0 {
		beerListText = "Пиво закончилось :("
	} else {
		for _, beer := range beersList {
			beerInfo := utils.FormatBeerInfo(beer, false) // false - краткая информация
			beerListText += fmt.Sprintf("%s\n\n", beerInfo)
		}
	}

	sendMessage(bot, message.Chat.ID, beerListText, "Markdown", nil)

}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	if waitingForSearchQuery[message.Chat.ID] {
		handleSearchMessage(bot, message, db)
		delete(waitingForSearchQuery, message.Chat.ID)
	} else {
		switch message.Text {
		case "Показать пиво":
			beersMutex.Lock()
			beersList := beers
			beersMutex.Unlock()
			var beerListText string

			if len(beersList) == 0 {
				beerListText = "Пиво закончилось :("
			} else {
				for _, beer := range beersList {
					beerInfo := utils.FormatBeerInfo(beer, false) // false - краткая информация
					beerListText += fmt.Sprintf("%s\n\n", beerInfo)
				}
			}

			sendMessage(bot, message.Chat.ID, beerListText, "Markdown", nil)
		case "Найти пиво":
			handleSearchCallback(bot, message) // Вызываем как callback, чтобы запустить поиск
		case "Корзина":
			handleCartCallback(bot, message, db)
		default:
			sendMessage(bot, message.Chat.ID, "Неизвестная команда.", "", nil)
		}
	}
}
