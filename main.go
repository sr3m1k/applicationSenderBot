package main

import (
	"applicationBot/botApp"
	"applicationBot/config"
	"applicationBot/database"
	"applicationBot/repoRequests"
	"applicationBot/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/pelletier/go-toml"
	"log"
	_ "os"
)

const (
	configPath = "config.toml"
)

func main() {
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}
	log.Printf("Конфиг успешно загружен")

	DB, err := database.InitDB(conf.Database.Path)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	log.Printf("Установлено подключение к базе данных")

	defer func() {
		err := DB.Close()
		if err != nil {
			log.Fatalf("Ошибка закрытия БД: %v", err)

		}
	}()

	requestRepo := repoRequests.NewRequestRepository(DB)
	requestService := service.NewRequestService(requestRepo)

	bot, err := tgbotapi.NewBotAPI(conf.Bot.Token)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}
	log.Printf("Бот успешно создан")

	botApp.StartBot(bot, DB, conf, requestRepo, requestService)
}
