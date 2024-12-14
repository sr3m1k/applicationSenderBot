package botApp

import (
	"applicationBot/config"
	"applicationBot/repoRequests"
	"applicationBot/service"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(bot *tgbotapi.BotAPI, db *sql.DB, config config.Config, requestRepo *repoRequests.RequestRepository, reqService *service.RequestService) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {

		if update.Message != nil {
			handleMessage(bot, db, update.Message, config, requestRepo, reqService)
		}
		if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery, config, reqService)
		}
	}
}
