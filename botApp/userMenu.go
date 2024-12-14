package botApp

import (
	"applicationBot/service"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"time"
)

func showMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить заявку", "add_request"),
			tgbotapi.NewInlineKeyboardButtonData("Все заявки", "all_requests"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Главное меню")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func getMainMenuButton() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
		),
	)
}

func showRequestDetails(bot *tgbotapi.BotAPI, chatID int64, number int, requestService *service.RequestService) {
	request, err := requestService.GetRequestDetails(number)
	if err != nil {
		sendMessage(bot, chatID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	text := fmt.Sprintf("Заявка %d\nКомментарий: %s", request.Number, request.Comment)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить таймер", fmt.Sprintf("add_timer_%d", number)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить заявку", fmt.Sprintf("del_%d", number)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назад к списку", "all_requests"),
			tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showRequestsList(chatID int64, userID int64, bot *tgbotapi.BotAPI, requestRepo *service.RequestService) {
	//rows, err := db.Query("SELECT number, comment FROM requests WHERE user_id = ? ORDER BY number", userID)

	requests, err := requestRepo.GetRequestsByUserId(userID)
	if err != nil {
		sendMessage(bot, chatID, "Ошибка при получении списка заявок")
		return
	}
	if len(requests) == 0 {
		sendMessage(bot, chatID, "У вас нет заявок")
		return
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, req := range requests {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", req.Number),
			fmt.Sprintf("req_%d", req.Number),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
	})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(chatID, "Ваши заявки:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleNumberInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	number, err := strconv.Atoi(message.Text)
	if err != nil {
		sendMessage(bot, chatID, "Введите корректный номер заявки")
		return
	}
	userStates[chatID] = "waiting_comment"
	tempData[chatID] = number

	sendMessage(bot, chatID, "Введите комментарий к заявке")
}

func handleCommentInput(bot *tgbotapi.BotAPI, message *tgbotapi.Message, requestService *service.RequestService) {
	chatID := message.Chat.ID
	number := tempData[chatID]

	err := requestService.CreateRequest(number, message.Text, message.Chat.ID, message.From.UserName, time.Now().Format("2006-01-02 15:04:05"))

	if err != nil {
		sendMessage(bot, chatID, "Ошибка при добавлении заявки")
		fmt.Println(err)
		return
	}

	delete(userStates, chatID)
	delete(tempData, chatID)

	sendMessage(bot, chatID, "Заявка успешно добавлена")
	showRequestDetails(bot, chatID, number, requestService)

}

func deleteRequest(bot *tgbotapi.BotAPI, chatID int64, number int, requestService *service.RequestService) {
	err := requestService.DeleteRequest(number)
	if err != nil {
		sendMessage(bot, chatID, "Ошибка при удалении заявки"+err.Error())
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Заявка %d успешно удалена", number))
	bot.Send(msg)
	getMainMenuButton()
}

func showTimerOptions(bot *tgbotapi.BotAPI, chatID int64, number int) {
	text := fmt.Sprintf("Выберите время для таймера заявки %d:", number)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("5 минут", fmt.Sprintf("timer_5_%d", number)),
			tgbotapi.NewInlineKeyboardButtonData("15 минут", fmt.Sprintf("timer_15_%d", number)),
			tgbotapi.NewInlineKeyboardButtonData("30 минут", fmt.Sprintf("timer_30_%d", number)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func setTimer(bot *tgbotapi.BotAPI, chatID int64, number, minutes int) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Таймер на %d минут установлен для заявки %d.", minutes, number))
	bot.Send(msg)

	go func(chatID int64, number, minutes int) {
		time.Sleep(time.Duration(minutes) * time.Minute)

		text := fmt.Sprintf("Напоминание! Проверьте заявку №%d.", number)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Добавить таймер", fmt.Sprintf("add_timer_%d", number)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Удалить заявку", fmt.Sprintf("del_%d", number)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
		}
	}(chatID, number, minutes)
}

func isAdmin(userID int64, Admins []int64) bool {
	for _, id := range Admins {
		if userID == id {
			return true
		}
	}
	return false
}

func isUserAllowed(userID int64, Users []int64) bool {
	for _, id := range Users {
		if userID == id {
			return true
		}
	}
	return false
}
