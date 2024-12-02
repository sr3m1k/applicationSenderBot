package botApp

import (
	"applicationBot/configuration"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(bot *tgbotapi.BotAPI, db *sql.DB, config configuration.Config) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {

		if update.Message != nil {
			handleMessage(bot, db, update.Message, config)
		}
		if update.CallbackQuery != nil {
			handleCallback(bot, db, update.CallbackQuery, config)
		}
	}
}
