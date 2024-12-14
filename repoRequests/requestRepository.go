package repoRequests

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type RequestRepository struct {
	db *sql.DB
}

func NewRequestRepository(db *sql.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

type Request struct {
	Number   int
	Comment  string
	UserId   int
	Username string
	Datetime string
}

func (r *RequestRepository) AddRequest(request Request) error {

	query := `INSERT INTO requests (number, comment, user_id, username, datetime) VALUES (?, ?, ?, ?,?)`

	_, err := r.db.Exec(query,
		request.Number,
		request.Comment,
		request.UserId,
		request.Username,
		request.Datetime)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("INSERT INTO NoDellRequests (number, comment, user_id, username, datetime) VALUES (?, ?, ?, ?,?)",
		request.Number,
		request.Comment,
		request.UserId,
		request.Username,
		request.Datetime)

	return err

}

func (r *RequestRepository) GetRequestsByUserId(userID int64) ([]Request, error) {

	rows, err := r.db.Query(`SELECT number, comment FROM requests WHERE user_id = ? ORDER BY number`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var request Request
		if err := rows.Scan(&request.Number, &request.Comment); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func (r *RequestRepository) GetRequestByNumber(number int) (Request, error) {
	var request Request

	row := r.db.QueryRow("SELECT number, comment FROM requests WHERE number = ?", number)

	err := row.Scan(&request.Number, &request.Comment)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Request{}, fmt.Errorf("Заявка с номером %d - не найдена ", number)
		}
		return Request{}, err
	}

	return request, nil
}

func (r *RequestRepository) DeleteRequestByNumber(number int) error {
	_, err := r.db.Exec("DELETE FROM requests WHERE number = ?", number)
	return err
}

func (r *RequestRepository) GetChatIds(table string) ([]int64, error) {
	query := fmt.Sprintf("SELECT chat_id FROM %s", table)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		chatIDs = append(chatIDs, chatID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chatIDs, nil
}

func (r *RequestRepository) AddChatToDB(table string, chatID int64, chatTitle string) error {
	query := fmt.Sprintf("INSERT OR IGNORE INTO %s (chat_id, chat_title) VALUES (?, ?)", table)
	result, err := r.db.Exec(query, chatID, chatTitle)
	if err != nil {
		log.Printf("Ошибка при добавлении чата в таблицу %s: %v", table, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка при проверке количества затронутых строк: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("Чат %d (%s) уже существует в таблице %s", chatID, chatTitle, table)
	} else {
		log.Printf("Чат %d (%s) успешно добавлен в таблицу %s", chatID, chatTitle, table)
	}

	return nil
}
