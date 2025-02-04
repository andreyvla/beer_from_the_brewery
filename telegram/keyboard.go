package telegram

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func createMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Показать пиво"),
			tgbotapi.NewKeyboardButton("Найти пиво"),
			tgbotapi.NewKeyboardButton("Корзина"),
		),
	)
}

func createQuantityKeyboard(beerID, quantity int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("-", fmt.Sprintf("adjust_quantity:%d:%d:-1", beerID, quantity-1)),
			tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(quantity), fmt.Sprintf("quantity:%d:%d", beerID, quantity)), // Текущее количество (неактивная кнопка)
			tgbotapi.NewInlineKeyboardButtonData("+", fmt.Sprintf("adjust_quantity:%d:%d:1", beerID, quantity+1)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", fmt.Sprintf("confirm_add:%d:%d", beerID, quantity)),
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
