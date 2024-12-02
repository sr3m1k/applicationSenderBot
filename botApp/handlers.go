package botApp

import (
	"applicationBot/configuration"
	"applicationBot/database"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

var (
	userStates      = make(map[int64]string)
	tempData        = make(map[int64]int)
	DB              *sql.DB
	broadcastTarget string
	//config           Config
	broadcastMessage string
)

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message, config configuration.Config) {

	if message.Text == "торг" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "ворк")
		bot.Send(msg)
	}

	chatID := message.Chat.ID
	userID := message.From.ID
	chatTitle := message.Chat.Title
	if message.Chat.Type == "private" {
		if !isUserAllowed(userID, config.Access.Users) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		}

	}
	log.Printf("Получено сообщение: ChatType=%s, ChatID=%d, UserID=%d, Username=%s, Text=%s",
		message.Chat.Type, chatID, userID, message.From.UserName, message.Text)

	if message.Text == "/trader" {
		if !isAdmin(userID, config.Access.Admins) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		}
		err := database.AddChatToDB(db, "traders", chatID, chatTitle)
		if err != nil {
			sendMessage(bot, chatID, "Ошибка при добавлении чата в бд")
		} else {
			sendMessage(bot, chatID, "Трейдер")
		}
		return
	}

	if message.Text == "/merch" {
		if !isAdmin(userID, config.Access.Admins) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		}
		err := database.AddChatToDB(db, "merchants", chatID, chatTitle)
		if err != nil {
			sendMessage(bot, chatID, "Ошибка при добавлении чата в бд")
		} else {
			sendMessage(bot, chatID, "Мерчант")
		}
		return
	}

	if message.Chat.IsPrivate() {
		if isUserAllowed(userID, config.Access.Users) {
			state, exists := userStates[chatID]

			if exists && state == "waiting_broadcast_message" {
				if !isAdmin(userID, config.Access.Admins) {
					sendMessage(bot, chatID, "Недостаточно прав")
					return
				}

				err := spamMessage(bot, db, broadcastTarget, message)
				if err != nil {
					sendMessage(bot, chatID, fmt.Sprintf("Ошибка при рассылке: %v", err))
				} else {
					sendMessage(bot, chatID, fmt.Sprintf("Сообщение успешно отправлено всем %s.", broadcastTarget))
				}

				delete(userStates, chatID)
				broadcastTarget = ""
				showMainMenu(bot, chatID)
				return
			}

			if message.Text == "/start" {
				showMainMenu(bot, chatID)
				return
			}

			if message.Text == "/admin" {
				if !isAdmin(userID, config.Access.Admins) {
					sendMessage(bot, chatID, "Недостаточно прав")
					return
				}
				showAdminMenu(bot, chatID)
				return
			}

			if !exists {
				showMainMenu(bot, chatID)
				return
			}

			switch state {
			case "waiting_number":
				handleNumberInput(bot, db, message)
			case "waiting_comment":
				handleCommentInput(bot, db, message)
			default:
				showMainMenu(bot, chatID)
			}
		}
	}
}

func handleCallback(bot *tgbotapi.BotAPI, db *sql.DB, callback *tgbotapi.CallbackQuery, config configuration.Config) {

	data := callback.Data
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID
	userID := callback.From.ID

	if callback.Message.Chat.Type != "private" {
		sendMessage(bot, chatID, "Этот бот работает только в личных чатах.")
		return
	}

	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	bot.Request(deleteMsg)

	switch {

	case data == "send_broadcast":
		if !isAdmin(userID, config.Access.Admins) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		}
		showBroadcastOptions(bot, chatID)
	//case data == "broadcast_traders":
	//	sendMessage(bot, chatID, "Введите сообщение для рассылки трейдерам:")
	//	spamMode = true
	//	spamTarget = "traders"
	//
	//	if spamMode {
	//		if spamTarget == "traders" {
	//			err := spamMessage(bot, db, "traders", callback.Message.Text)
	//			if err != nil {
	//				sendMessage(bot, chatID, fmt.Sprintf("Ошибка при рассылке: %v", err))
	//			} else {
	//				sendMessage(bot, chatID, "Сообщение успешно отправлено всем трейдерам.")
	//			}
	//		}
	//
	//		spamMode = false
	//		spamTarget = ""
	//		return
	//	}
	//case data == "broadcast_merchants":
	//	sendMessage(bot, chatID, "Введите сообщение для рассылки мерчам:")

	case data == "broadcast_traders":
		if !isAdmin(userID, config.Access.Admins) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		}
		broadcastTarget = "traders"
		sendMessage(bot, chatID, "Введите сообщение для рассылки трейдерам:")
		userStates[chatID] = "waiting_broadcast_message"

	case data == "broadcast_merchants":
		if !isAdmin(userID, config.Access.Admins) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		}
		broadcastTarget = "merchants"
		sendMessage(bot, chatID, "Введите сообщение для рассылки мерчантам:")
		userStates[chatID] = "waiting_broadcast_message"

	case data == "add_request":
		userStates[chatID] = "waiting_number"
		msg := tgbotapi.NewMessage(chatID, "Введите номер заявки:")
		msg.ReplyMarkup = getMainMenuButton()
		bot.Send(msg)

	case data == "all_requests":
		showRequestsList(chatID, userID, db, bot)

	case data == "main_menu":
		showMainMenu(bot, chatID)

	case data[:9] == "add_timer":
		parts := strings.Split(data, "_")
		if len(parts) != 3 {
			sendMessage(bot, chatID, "Некорректный запрос для добавления таймера")
			return
		}

		number, err := strconv.Atoi(parts[2])
		if err != nil {
			sendMessage(bot, chatID, "Некорректный номер заявки")
			return
		}

		showTimerOptions(bot, chatID, number)

	case data[:6] == "timer_":
		parts := strings.Split(data, "_")
		if len(parts) != 3 {
			sendMessage(bot, chatID, "Некорректный запрос для таймера")
			return
		}
		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			sendMessage(bot, chatID, "Некорректное время таймера")
			return
		}

		number, err := strconv.Atoi(parts[2])
		if err != nil {
			sendMessage(bot, chatID, "Некорректный номер заявки")
			return
		}

		setTimer(bot, chatID, number, minutes)

	case len(data) > 4 && data[:4] == "req_":
		numberStr := data[4:]
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			sendMessage(bot, chatID, "Некорректный номер заявки.")
			return
		}
		showRequestDetails(bot, chatID, number, db)

	case len(data) > 4 && data[:4] == "del_":
		numberStr := data[4:]
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			sendMessage(bot, chatID, "Некорректный номер заявки.")
			return
		}
		deleteRequest(db, bot, chatID, number)
		showMainMenu(bot, chatID)

	}

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	bot.Request(callbackConfig)
}

func spamMessage(bot *tgbotapi.BotAPI, db *sql.DB, table string, originalMessage *tgbotapi.Message) error {
	rows, err := db.Query(fmt.Sprintf("SELECT chat_id FROM %s", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return err
		}

		copyMsg := tgbotapi.NewCopyMessage(chatID, originalMessage.Chat.ID, originalMessage.MessageID)
		_, err = bot.Send(copyMsg)
		if err != nil {
			log.Printf("Не удалось отправить сообщение в чат %d: %v", chatID, err)
		}
	}

	return nil
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
	}
}
