package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pelletier/go-toml"
)

type Config struct {
	Bot struct {
		Token string `toml:"token"`
	} `toml:"bot"`
	Access struct {
		Users  []int64 `toml:"users"`
		Admins []int64 `toml:"admins"`
	} `toml:"access"`
}

var (
	userStates = make(map[int64]string)
	tempData   = make(map[int64]int)
	DB         *sql.DB
	spamMode   bool
	spamTarget string
	config     Config
)

func loadConfig() error {
	file, err := os.Open("config.toml")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	return decoder.Decode(&config)
}

func main() {
	err := loadConfig()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	DB, err = initDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	} else {
		log.Printf("Установлено подключение к базе данных")
	}
	defer DB.Close()

	bot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	} else {
		log.Printf("Бот успешно создан")
	}

	startBot(bot, DB)
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./data/requests.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS requests (
            number INTEGER PRIMARY KEY,
            comment TEXT,
            user_id INTEGER,
            username TEXT,
            datetime TEXT                               
        );
        CREATE TABLE IF NOT EXISTS NoDellRequests (
            number INTEGER PRIMARY KEY,
            comment TEXT,
            user_id INTEGER,
            username TEXT,
            datetime TEXT                               
        );
        CREATE TABLE IF NOT EXISTS traders (
    		chat_id INTEGER PRIMARY KEY,
    		chat_title TEXT
		);
		CREATE TABLE IF NOT EXISTS merchants (
    		chat_id INTEGER PRIMARY KEY,
    		chat_title TEXT
		);

    `)
	return db, err
}

func startBot(bot *tgbotapi.BotAPI, db *sql.DB) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, db, update.Message)
		}
		if update.CallbackQuery != nil {
			handleCallback(bot, db, update.CallbackQuery)
		}
	}
}

func isAdmin(userID int64) bool {
	for _, id := range config.Access.Admins {
		if userID == id {
			return true
		}
	}
	return false
}

func isUserAllowed(userID int64) bool {
	for _, id := range config.Access.Users {
		if userID == id {
			return true
		}
	}
	return false
}

func handleMessage(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID
	chatTitle := message.Chat.Title

	if message.Text == "/trader" {
		if !isAdmin(userID) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		} else {
			err := addChatToDB(db, "traders", chatID, chatTitle)
			if err != nil {
				sendMessage(bot, chatID, "Ошибка при добавлении чата в бд")
			} else {
				sendMessage(bot, chatID, "Трейдер")
			}
			return
		}
	}

	if message.Text == "/merch" {
		if !isAdmin(userID) {
			sendMessage(bot, chatID, "Недостаточно прав")
			return
		} else {
			err := addChatToDB(db, "merchants", chatID, chatTitle)
			if err != nil {
				sendMessage(bot, chatID, "Ошибка при добавлении чата в бд")
			} else {
				sendMessage(bot, chatID, "Мерчант")
			}
			return
		}
	}

	if message.Chat.IsPrivate() {
		if isUserAllowed(userID) {
			//if message.Text == "/spam_trader" {
			//	spamMode = true
			//	spamTarget = "traders"
			//	sendMessage(bot, chatID, "Введите текст сообщения для рассылки трейдерам:")
			//	return
			//}
			//
			//if spamMode && spamTarget == "traders" {
			//	err := spamMessage(bot, db, "traders", message.Text)
			//	if err != nil {
			//		sendMessage(bot, chatID, fmt.Sprintf("Ошибка при рассылке: %v", err))
			//	} else {
			//		sendMessage(bot, chatID, "Сообщение успешно отправлено всем трейдерам.")
			//	}
			//	spamMode = false
			//	spamTarget = ""
			//	return
			//}

			if message.Text == "/start" {
				showMainMenu(bot, chatID)
				return
			}

			state, exists := userStates[chatID]
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

func handleCallback(bot *tgbotapi.BotAPI, db *sql.DB, callback *tgbotapi.CallbackQuery) {

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
	showMainMenu(bot, chatID)
}

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

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
	}
}

func addChatToDB(db *sql.DB, table string, chatID int64, chatTitle string) error {
	query := fmt.Sprintf("INSERT OR IGNORE INTO %s (chat_id, chat_title) VALUES (?, ?)", table)
	_, err := db.Exec(query, chatID, chatTitle)
	return err
}

func spamMessage(bot *tgbotapi.BotAPI, db *sql.DB, table string, messageText string) error {
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

		msg := tgbotapi.NewMessage(chatID, messageText)
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Не удалось отправить сообщение в чат %d: %v", chatID, err)
		}
	}

	return nil
}
