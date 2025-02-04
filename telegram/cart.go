package telegram

import (
	"beer_from_the_brewery/database"
	"beer_from_the_brewery/models"
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handleCartCallback(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *sql.DB) {
	loadCart, ok := carts.Load(message.Chat.ID) // Используем message.Chat.ID

	if !ok || loadCart == nil {
		sendMessage(bot, message.Chat.ID, "Ваша корзина пуста.", "", nil)
		return
	}

	cart := loadCart.(map[int]models.CartItem)

	if len(cart) == 0 {
		sendMessage(bot, message.Chat.ID, "Ваша корзина пуста.", "", nil)
		return
	}

	var cartText string
	var totalPrice float64

	for beerID, cartItem := range cart {
		beer, err := database.GetBeerByID(db, beerID)
		if err != nil {
			sendMessage(bot, message.Chat.ID, "Ошибка при получении данных о пиве.", "", nil)
			return
		}
		if beer == nil {
			sendMessage(bot, message.Chat.ID, "Пиво не найдено.", "", nil)
			return
		}
		beerPrice := beer.Price * float64(cartItem.Quantity)
		cartText += fmt.Sprintf("*%s*\nКоличество: %d\nЦена: %.2f\n\n", beer.Name, cartItem.Quantity, beerPrice)
		totalPrice += beerPrice
	}

	cartText += fmt.Sprintf("\nОбщая стоимость: %.2f", totalPrice)

	keyboard := createCartKeyboard() // Создаем клавиатуру для действий с корзиной
	sendMessage(bot, message.Chat.ID, cartText, "Markdown", &keyboard)

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
