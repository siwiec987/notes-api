package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func DoRecordsExistForUser(db *sql.DB, table string, userID int, recordIDs []int) (bool, error) {
	placeholders, args := BuildQueryArgs(userID, recordIDs)

	query := fmt.Sprintf(`
    SELECT COUNT(*) FROM %s 
    WHERE user_id = ? AND id IN (%s)`, table, placeholders)

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count == len(recordIDs), err
}

func WithTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New("failed to commit transaction")
	}

	return nil
}

func BuildQueryArgs(userID int, ids []int) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids)+1)
	args[0] = userID
	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	return strings.Join(placeholders, ","), args
}
