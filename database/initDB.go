package database

import (
	"database/sql"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitDB(DBPath string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		return nil, err
	}

	return db, err
}
