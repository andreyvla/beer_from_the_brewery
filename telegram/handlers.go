package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"beer_from_the_brewery/utils"
	"context"
	"fmt"
	"log"
	"strconv"

	"database/sql"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// handleCommand обрабатывает команды, отправленные боту.
func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB, logger *log.Logger) {
	switch message.Command() {
	case "start":
		handleStartCommand(bot, message)
	default:
		sendMessage(bot, message.Chat.ID, "Неизвестная команда.", "", nil, logger)
	}
}

// handleCallbackQuery обрабатывает callback-запросы от inline-клавиатур.
func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, logger *log.Logger) {
	switch {
	case strings.HasPrefix(callbackQuery.Data, "add_to_cart:"):
		handleAddToCartCallback(bot, callbackQuery, db, logger)
	case strings.HasPrefix(callbackQuery.Data, "adjust_quantity:"):
		handleAdjustQuantityCallback(bot, callbackQuery, db, logger)
	case strings.HasPrefix(callbackQuery.Data, "confirm_add:"):
		handleConfirmAddCallback(bot, callbackQuery, db, logger)
	case callbackQuery.Data == "checkout":
		handleCheckoutCallback(bot, callbackQuery, db, logger)
	case callbackQuery.Data == "clear_cart":
		handleClearCartCallback(bot, callbackQuery, logger)
	case callbackQuery.Data == "beer":
		handleBeerCallback(bot, callbackQuery.Message, db, logger)
	case callbackQuery.Data == "search":
		handleSearchCallback(bot, callbackQuery.Message)
	case callbackQuery.Data == "cart":
		handleCartCallback(bot, callbackQuery.Message, db, logger)

	default:
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неизвестное действие.", "", nil, logger)
	}
}

// handleStartCommand обрабатывает команду /start.
func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот для покупки пива.")
	msg.ReplyMarkup = createMainKeyboard()
	bot.Send(msg)
}

// handleAddToCartCallback обрабатывает callback-запрос на добавление пива в корзину.
func handleAddToCartCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, logger *log.Logger) {
	data := strings.Split(callbackQuery.Data, ":")
	if len(data) != 3 {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат данных.", "", nil, logger)
		return
	}

	beerID, err := strconv.Atoi(data[1])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный ID пива.", "", nil, logger)
		return
	}

	beer, err := database.GetBeerByID(context.Background(), db, beerID)
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil, logger)
		return
	}

	if beer == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Пиво не найдено.", "", nil, logger)
		return
	}

	keyboard := createQuantityKeyboard(beerID, 1)

	sendMessage(bot, callbackQuery.Message.Chat.ID, fmt.Sprintf("Укажите количество %s:", beer.Name), "", &keyboard, logger)

}

// handleAdjustQuantityCallback обрабатывает callback-запрос на изменение количества пива в корзине.
func handleAdjustQuantityCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, logger *log.Logger) {

	data := strings.Split(callbackQuery.Data, ":")
	if len(data) != 4 {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат данных.", "", nil, logger)
		return
	}
	beerID, err := strconv.Atoi(data[1])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный ID пива.", "", nil, logger)
		return
	}
	quantity, err := strconv.Atoi(data[2])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат количества.", "", nil, logger)
		return
	}
	adjust, err := strconv.Atoi(data[3])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при изменении количества.", "", nil, logger)
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
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при обновлении сообщения", "", nil, logger)
	}
}

// handleConfirmAddCallback обрабатывает callback-запрос на подтверждение добавления пива в корзину.
func handleConfirmAddCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, logger *log.Logger) {
	data := strings.Split(callbackQuery.Data, ":")
	beerID, err := strconv.Atoi(data[1])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный ID пива.", "", nil, logger)
		return
	}
	quantity, err := strconv.Atoi(data[2])
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Неверный формат количества.", "", nil, logger)
		return
	}
	beer, err := database.GetBeerByID(context.Background(), db, beerID)
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil, logger)
		return
	}
	if beer == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Пиво не найдено.", "", nil, logger)
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

	sendMessage(bot, callbackQuery.Message.Chat.ID, fmt.Sprintf("%s (%d шт.) добавлен в корзину.", beer.Name, quantity), "", nil, logger)
}

// handleBeerCallback обрабатывает команду "Показать пиво".
func handleBeerCallback(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB, logger *log.Logger) {
	beersMutex.Lock()
	beersList := beers
	beersMutex.Unlock()
	var beerListText string

	if len(beersList) == 0 {
		beerListText = "Пиво закончилось :("
	} else {
		for _, beer := range beersList {
			beerInfo := utils.FormatBeerInfo(beer, false)
			beerListText += fmt.Sprintf("%s\n\n", beerInfo)
		}
	}

	sendMessage(bot, message.Chat.ID, beerListText, "Markdown", nil, logger)

}

// handleMessage обрабатывает сообщения, не являющиеся командами.
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB, logger *log.Logger) {
	if waitingForSearchQuery[message.Chat.ID] {
		handleSearchMessage(bot, message, db, logger)
		delete(waitingForSearchQuery, message.Chat.ID)
	} else {
		switch message.Text {
		case "Показать пиво":
			handleBeerCallback(bot, message, db, logger)
		case "Найти пиво":
			handleSearchCallback(bot, message)
		case "Корзина":
			handleCartCallback(bot, message, db, logger)
		default:
			sendMessage(bot, message.Chat.ID, "Неизвестная команда.", "", nil, logger)
		}
	}
}
