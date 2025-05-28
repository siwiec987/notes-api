package migrations

import (
	"database/sql"
	"fmt"
)

func InitializeTables(db *sql.DB) error {
	userTable :=  `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(50) UNIQUE NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL
        );
    `

	categoriesTable :=  `
        CREATE TABLE IF NOT EXISTS categories (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) UNIQUE NOT NULL,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
        );
    `

	notesTable := `
        CREATE TABLE IF NOT EXISTS notes (
            id SERIAL PRIMARY KEY,
            content TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE
        );
    `

	tables := []string {userTable, categoriesTable, notesTable}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}