package botApp

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func showRequestDetails(bot *tgbotapi.BotAPI, chatID int64, number int, db *sql.DB) {
	var comment string
	err := db.QueryRow("SELECT comment FROM requests WHERE number = ?", number).Scan(&comment)
	if err != nil {
		sendMessage(bot, chatID, "Заявка не найдена")
		return
	}

	text := fmt.Sprintf("Заявка %d\nКомментарий: %s", number, comment)

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

func showRequestsList(chatID int64, userID int64, db *sql.DB, bot *tgbotapi.BotAPI) {
	rows, err := db.Query("SELECT number, comment FROM requests WHERE user_id = ? ORDER BY number", userID)
	if err != nil {
		sendMessage(bot, chatID, "Ошибка при получении списка заявок")
		return
	}
	defer rows.Close()

	var buttons [][]tgbotapi.InlineKeyboardButton
	for rows.Next() {
		var number int
		var comment string
		rows.Scan(&number, &comment)

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", number),
			fmt.Sprintf("req_%d", number),
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

func handleNumberInput(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
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

func handleCommentInput(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	number := tempData[chatID]

	_, err := db.Exec("INSERT INTO requests (number, comment, user_id, username, datetime) VALUES (?, ?, ?, ?,?)",
		number, message.Text, message.Chat.ID, message.From.UserName, time.Now().Format("2006-01-02 15:04:05"))
	_, err = db.Exec("INSERT INTO NoDellRequests (number, comment, user_id, username, datetime) VALUES (?, ?, ?, ?, ?)",
		number, message.Text, message.Chat.ID, message.From.UserName, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		sendMessage(bot, chatID, "Ошибка при добавлении заявки")
		fmt.Println(err)
		return
	}

	delete(userStates, chatID)
	delete(tempData, chatID)

	sendMessage(bot, chatID, "Заявка успешно добавлена")
	showRequestDetails(bot, chatID, number, db)
	//showMainMenu(bot, chatID)
}

func deleteRequest(db *sql.DB, bot *tgbotapi.BotAPI, chatID int64, number int) {
	_, err := db.Exec("DELETE FROM requests WHERE number = ?", number)
	if err != nil {
		sendMessage(bot, chatID, "Ошибка при удалении заявки")
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Заявка %d успешно удалена", number))
	bot.Send(msg)
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
		bot.Send(msg)
	}(chatID, number, minutes)
}
