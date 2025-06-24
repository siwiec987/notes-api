package validation

import (
	"errors"
	"fmt"
	"time"
)

func isDateCorrect(s string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", s)
	return err == nil
}

func createDateFilter(query *string, args *[]any, param, operator string) error {
	if param != "" {
		if !isDateCorrect(param) {
			return errors.New("invalid date format, example: 2025-06-23 22:49:00")
		}

		*query += fmt.Sprintf(" AND %s ?", operator)
		*args = append(*args, param)
	}
	return nil
}

func CreateDateFilters(query *string, args *[]any, paramOperatorMap map[string]string) error {
	for param, operator := range paramOperatorMap {
		err := createDateFilter(query, args, param, operator)
		if err != nil {
			return err
		}
	}

	return nil
}