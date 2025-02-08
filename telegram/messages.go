package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// sendMessage отправляет сообщение в Telegram.
//
// bot - указатель на экземпляр бота.
// chatID - ID чата, куда нужно отправить сообщение.
// text - текст сообщения.
// parseMode - режим парсинга текста
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string, parseMode string, keyboard *tgbotapi.InlineKeyboardMarkup, logger *log.Logger) {
	msg := tgbotapi.NewMessage(chatID, text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	if keyboard != nil {
		msg.ReplyMarkup = *keyboard // Разыменовываем указатель для прикрепления клавиатуры.

	}
	if _, err := bot.Send(msg); err != nil {
		logger.Printf("Ошибка при отправке сообщения: %s", err.Error())
	}
}
