package main

import (
	"applicationBot/botApp"
	"applicationBot/configuration"
	"applicationBot/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/pelletier/go-toml"
	"log"
	_ "os"
)

func main() {
	config, err := configuration.LoadConfig()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	} else {
		log.Printf("Конфиг успешно загружен")
	}

	DB, err := database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	} else {
		log.Printf("Установлено подключение к базе данных")
	}
	defer func() {
		err := DB.Close()
		if err != nil {
			log.Fatalf("Ошибка закрытия БД: %v", err)

		}
	}()

	bot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	} else {
		log.Printf("Бот успешно создан")
	}

	botApp.StartBot(bot, DB, config)
}
