package database

import (
	"database/sql"
	"fmt"
	"os"

	_"github.com/go-sql-driver/mysql"
	"github.com/siwiec987/notes-api/internal/migrations"
)

var (
	host = os.Getenv("DB_HOST")
	port = os.Getenv("DB_PORT")
	user = os.Getenv("DB_USER")
	pass = os.Getenv("DB_PASS")
	name = os.Getenv("DB_NAME")
)

func GetDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, err
}

func DBInit(db *sql.DB) error {
	err := migrations.InitializeTables(db)
	return err
}