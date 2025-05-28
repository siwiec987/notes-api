package migrations

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func SeedData(db *sql.DB) error {
	// Przykładowi użytkownicy
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	_, err = db.Exec(fmt.Sprintf("INSERT INTO users (username, email, password) VALUES ('admin', 'admin@example.com', '%s')", hashedPassword))
	// 	`
	// 	INSERT INTO users (username, email, password) VALUES
	// 	('admin', 'admin@example.com', 'admin')
	// `)
	if err != nil {
		return fmt.Errorf("failed to seed users: %v", err)
	}

	// Przykładowe kategorie
	_, err = db.Exec(`
		INSERT INTO categories (name, user_id) VALUES
		('Personal', 1),
		('Work', 1)
	`)
	if err != nil {
		return fmt.Errorf("failed to seed categories: %v", err)
	}

	// Przykładowe notatki
	_, err = db.Exec(`
		INSERT INTO notes (content, user_id, category_id) VALUES
		('Buy milk', 1, 1),
		('Finish project', 1, 2),
		('Test', 1, 2)
	`)
	if err != nil {
		return fmt.Errorf("failed to seed notes: %v", err)
	}

	return nil
}
