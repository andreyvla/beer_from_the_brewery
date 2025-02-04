package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

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
