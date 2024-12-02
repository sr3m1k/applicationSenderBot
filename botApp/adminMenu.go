package botApp

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func showAdminMenu(bot *tgbotapi.BotAPI, id int64) {
	msg := tgbotapi.NewMessage(id, "Выберите действие:")
	msg.ReplyMarkup = getAdminMenuButton()
	bot.Send(msg)

}

func getAdminMenuButton() interface{} {

	mainButton := tgbotapi.NewInlineKeyboardButtonData("Сделать рассылку", "send_broadcast")

	//traderButton := tgbotapi.NewInlineKeyboardButtonData("Рассылка трейдерам", "broadcast_traders")
	//merchantButton := tgbotapi.NewInlineKeyboardButtonData("Рассылка мерчам", "broadcast_merchants")

	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{mainButton},
	)
}

func showBroadcastOptions(bot *tgbotapi.BotAPI, chatID int64) {
	text := fmt.Sprintf("Выберите кому сделать рассылку")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Трейдеры", fmt.Sprintf("broadcast_traders")),
			tgbotapi.NewInlineKeyboardButtonData("Мерчанты", fmt.Sprintf("broadcast_traders")),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
