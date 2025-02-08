package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"context"
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// handleCartCallback обрабатывает команду /cart, отображая содержимое корзины пользователя.
func handleCartCallback(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB, logger *log.Logger) {
	loadCart, ok := carts.Load(message.Chat.ID)

	if !ok || loadCart == nil {
		sendMessage(bot, message.Chat.ID, "Ваша корзина пуста.", "", nil, logger)
		return
	}

	cart := loadCart.(map[int]models.CartItem)

	if len(cart) == 0 {
		sendMessage(bot, message.Chat.ID, "Ваша корзина пуста.", "", nil, logger)
		return
	}

	var cartText string
	var totalPrice float64

	for beerID, cartItem := range cart {
		beer, err := database.GetBeerByID(context.Background(), db, beerID)
		if err != nil {
			logger.Printf("Ошибка при получении данных о пиве (ID: %d): %s", beerID, err.Error())
			sendMessage(bot, message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil, logger)
			return
		}
		if beer == nil {
			sendMessage(bot, message.Chat.ID, "Пиво не найдено.", "", nil, logger)
			return
		}
		beerPrice := beer.Price * float64(cartItem.Quantity)
		cartText += fmt.Sprintf("*%s*\nКоличество: %d\nЦена: %.2f\n\n", beer.Name, cartItem.Quantity, beerPrice)
		totalPrice += beerPrice
	}

	cartText += fmt.Sprintf("\nОбщая стоимость: %.2f", totalPrice)

	keyboard := createCartKeyboard() // Создаем клавиатуру для действий с корзиной
	sendMessage(bot, message.Chat.ID, cartText, "Markdown", &keyboard, logger)

}

// handleCheckoutCallback обрабатывает callback-запрос на оформление заказа.
func handleCheckoutCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, logger *log.Logger) {
	loadCart, ok := carts.Load(callbackQuery.Message.Chat.ID)
	if !ok || loadCart == nil {
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ваша корзина пуста. Нечего оформлять.", "", nil, logger)
		return
	}
	cart := loadCart.(map[int]models.CartItem)
	cartItems := make([]models.CartItem, 0, len(cart)) // Преобразуем map в slice для передачи в CreateOrder

	for _, cartItem := range cart {
		cartItems = append(cartItems, cartItem)
	}

	err := database.CreateOrder(context.Background(), db, callbackQuery.Message.Chat.ID, cartItems)
	if err != nil {
		logger.Printf("Ошибка при оформлении заказа (ChatID: %d): %s", callbackQuery.Message.Chat.ID, err.Error())
		sendMessage(bot, callbackQuery.Message.Chat.ID, "Ошибка при оформлении заказа. Пожалуйста, попробуйте позже.", "", nil, logger)
		return
	}

	carts.Delete(callbackQuery.Message.Chat.ID) // Очищаем корзину после успешного заказа
	keyboard := createBeerKeyboard()
	sendMessage(bot, callbackQuery.Message.Chat.ID, "Спасибо за ваш заказ!", "", &keyboard, logger)

}

// handleClearCartCallback обрабатывает callback-запрос на очистку корзины.
func handleClearCartCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, logger *log.Logger) {

	carts.Delete(callbackQuery.Message.Chat.ID)
	sendMessage(bot, callbackQuery.Message.Chat.ID, "Корзина очищена.", "", nil, logger)
}
