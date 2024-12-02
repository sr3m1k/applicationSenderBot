package database

import (
	"database/sql"
	"fmt"
	"log"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "/data/requests.db")
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

func AddChatToDB(db *sql.DB, table string, chatID int64, chatTitle string) error {
	query := fmt.Sprintf("INSERT OR IGNORE INTO %s (chat_id, chat_title) VALUES (?, ?)", table)
	result, err := db.Exec(query, chatID, chatTitle)
	if err != nil {
		log.Printf("Ошибка при добавлении чата в таблицу %s: %v", table, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка при проверке количества затронутых строк: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("Чат %d (%s) уже существует в таблице %s", chatID, chatTitle, table)
	} else {
		log.Printf("Чат %d (%s) успешно добавлен в таблицу %s", chatID, chatTitle, table)
	}

	return nil
}
