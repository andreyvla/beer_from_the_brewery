package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"beer_from_the_brewery/utils"
	"fmt"
	"strconv"
	"sync"

	"database/sql"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var beers []models.Beer
var beersMutex = &sync.Mutex{}
var waitingForSearchQuery = make(map[int64]bool)
var carts sync.Map // sync.Map[int64]map[int]models.CartItem  (для ясности)

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

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	switch message.Command() {
	case "start":
		handleStartCommand(bot, message)
	default:
		sendMessage(bot, message.Chat.ID, "Неизвестная команда.", "", nil)
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Показать пиво", "beer"),
			tgbotapi.NewInlineKeyboardButtonData("Найти пиво", "search"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Привет! Я бот для покупки пива. Выберите действие:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	switch callbackQuery.Data { // Используем callbackQuery.Data
	case "beer":
		handleBeerCallback(bot, callbackQuery, db)
	case "search":
		handleSearchCallback(bot, callbackQuery)
	case "cart":
		handleCartCallback(bot, callbackQuery, db)
	case "checkout":
		handleCheckoutCallback(bot, callbackQuery, db)
	case "clear_cart":
		handleClearCartCallback(bot, callbackQuery)
	default:
		if strings.HasPrefix(callbackQuery.Data, "add_to_cart:") {
			handleAddToCartCallback(bot, callbackQuery, db)
		} else {

			sendMessage(bot, callbackQuery.Message.Chat.ID, "Неизвестное действие.", "", nil)
		}

	}
}
func handleBeerCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	beersMutex.Lock()   // Добавляем блокировку
	beersList := beers  // Создаем копию
	beersMutex.Unlock() // Освобождаем блокировку
	var beerListText string

	if len(beersList) == 0 {
		beerListText = "Пиво закончилось :(" // Выводим сообщение, если пива нет
	} else {
		for _, beer := range beersList {
			beerInfo := utils.FormatBeerInfo(beer, false) // false - краткая информация
			beerListText += fmt.Sprintf("%s\n\n", beerInfo)
		}
	}

	// Добавляем кнопки для добавления пива в корзину, используя ID пива
	var beerRows [][]tgbotapi.InlineKeyboardButton
	for _, beer := range beersList {
		// Используем NewInlineKeyboardRow для создания новой строки
		beerRows = append(beerRows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Добавить в корзину %s", beer.Name), fmt.Sprintf("add_to_cart:%d", beer.ID)),
		})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(beerRows...)

	sendMessage(bot, callbackQuery.Message.Chat.ID, beerListText, "Markdown", &keyboard)

}
func handleAddToCartCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	data := strings.Split(callbackQuery.Data, ":")
	if len(data) != 2 {
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

	loadCart, _ := carts.LoadOrStore(callbackQuery.Message.Chat.ID, make(map[int]models.CartItem)) // Загружаем корзину или создаем новую, если ее нет
	cart := loadCart.(map[int]models.CartItem)

	cart[beerID] = models.CartItem{BeerID: beerID, Quantity: cart[beerID].Quantity + 1} // Добавляем в корзину
	carts.Store(callbackQuery.Message.Chat.ID, cart)

	sendMessage(bot, callbackQuery.Message.Chat.ID, fmt.Sprintf("%s добавлено в корзину.", beer.Name), "", nil)

}

func handleCartCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	loadCart, ok := carts.Load(callbackQuery.Message.Chat.ID)

	if !ok || loadCart == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ваша корзина пуста.", "", nil)
		return
	}

	cart := loadCart.(map[int]models.CartItem)

	if len(cart) == 0 {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ваша корзина пуста.", "", nil)
		return
	}

	var cartText string
	var totalPrice float64

	for beerID, cartItem := range cart {
		beer, err := database.GetBeerByID(db, beerID)
		if err != nil {
			sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil)
			return
		}
		if beer == nil {
			sendMessage(bot, callbackQuery.Message.Chat.ID, "Пиво не найдено.", "", nil)
			return
		}
		beerPrice := beer.Price * float64(cartItem.Quantity)
		cartText += fmt.Sprintf("*%s*\nКоличество: %d\nЦена: %.2f\n\n", beer.Name, cartItem.Quantity, beerPrice)
		totalPrice += beerPrice
	}

	cartText += fmt.Sprintf("\nОбщая стоимость: %.2f", totalPrice)

	keyboard := createCartKeyboard() // Создаем клавиатуру для действий с корзиной
	sendMessage(bot, callbackQuery.Message.Chat.ID, cartText, "Markdown", &keyboard)

}

func handleCheckoutCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	loadCart, ok := carts.Load(callbackQuery.Message.Chat.ID)
	if !ok || loadCart == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ваша корзина пуста. Нечего оформлять.", "", nil)
		return
	}
	cart := loadCart.(map[int]models.CartItem)

	cartItems := make([]models.CartItem, 0, len(cart))
	for _, cartItem := range cart {
		cartItems = append(cartItems, cartItem)
	}

	err := database.CreateOrder(db, callbackQuery.Message.Chat.ID, cartItems)
	if err != nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при оформлении заказа. Пожалуйста, попробуйте позже.", "", nil)
		return
	}

	carts.Delete(callbackQuery.Message.Chat.ID) // Очищаем корзину после успешного заказа
	keyboard := createBeerKeyboard()
	sendMessage(bot, callbackQuery.Message.Chat.ID, "Спасибо за ваш заказ!", "", &keyboard)

}

func handleClearCartCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {

	carts.Delete(callbackQuery.Message.Chat.ID)
	sendMessage(bot, callbackQuery.Message.Chat.ID, "Корзина очищена.", "", nil)
}

func handleSearchCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Введите название пива для поиска:")
	bot.Send(msg)
	waitingForSearchQuery[callbackQuery.Message.Chat.ID] = true
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
	}
	if len(foundBeers) == 1 {
		beer := foundBeers[0]

		msgText := utils.FormatBeerInfo(beer, true)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, msgText))
		return
	}

	var beerListText string
	for _, beer := range foundBeers {
		beerInfo := utils.FormatBeerInfo(beer, false)
		beerListText += fmt.Sprintf("%s\n\n", beerInfo)
	}

	sendMessage(bot, message.Chat.ID, beerListText, "", nil)
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	if waitingForSearchQuery[message.Chat.ID] {
		handleSearchMessage(bot, message, db)
		delete(waitingForSearchQuery, message.Chat.ID)
	}
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string, parseMode string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	if keyboard != nil {
		msg.ReplyMarkup = *keyboard // Разыменовываем указатель
	}
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %s", err.Error())
	}
}

func createBeerKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Показать пиво", "beer"),
			tgbotapi.NewInlineKeyboardButtonData("Найти пиво", "search"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Перейти к корзине", "cart"),
		),
	)
}

func createCartKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Оформить заказ", "checkout"),
			tgbotapi.NewInlineKeyboardButtonData("Очистить корзину", "clear_cart"),
		),
	)
}
