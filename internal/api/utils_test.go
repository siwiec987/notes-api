package api

import (
	"reflect"
	"testing"
)

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name         string
		limitStr     string
		defaultLimit int
		expected     int
	}{
		{
			name:         "Valid positive limit",
			limitStr:     "50",
			defaultLimit: 20,
			expected:     50,
		},
		{
			name:         "Empty string uses default",
			limitStr:     "",
			defaultLimit: 20,
			expected:     20,
		},
		{
			name:         "Invalid string uses default",
			limitStr:     "abc",
			defaultLimit: 20,
			expected:     20,
		},
		{
			name:         "Zero limit uses default",
			limitStr:     "0",
			defaultLimit: 20,
			expected:     20,
		},
		{
			name:         "Negative limit uses default",
			limitStr:     "-5",
			defaultLimit: 20,
			expected:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLimit(tt.limitStr, tt.defaultLimit)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParseOffset(t *testing.T) {
	tests := []struct {
		name          string
		offsetStr     string
		defaultOffset int
		expected      int
	}{
		{
			name:          "Valid positive offset",
			offsetStr:     "10",
			defaultOffset: 0,
			expected:      10,
		},
		{
			name:          "Empty string uses default",
			offsetStr:     "",
			defaultOffset: 0,
			expected:      0,
		},
		{
			name:          "Invalid string uses default",
			offsetStr:     "xyz",
			defaultOffset: 0,
			expected:      0,
		},
		{
			name:          "Zero offset uses default",
			offsetStr:     "0",
			defaultOffset: 5,
			expected:      5,
		},
		{
			name:          "Negative offset uses default",
			offsetStr:     "-3",
			defaultOffset: 0,
			expected:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOffset(tt.offsetStr, tt.defaultOffset)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestIsDateCorrect(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		expected bool
	}{
		{
			name:     "Valid date format",
			dateStr:  "2025-06-23 22:49:00",
			expected: true,
		},
		{
			name:     "Valid date format with zeros",
			dateStr:  "2025-01-01 00:00:00",
			expected: true,
		},
		{
			name:     "Invalid date format - missing time",
			dateStr:  "2025-06-23",
			expected: false,
		},
		{
			name:     "Invalid date format - wrong separator",
			dateStr:  "2025/06/23 22:49:00",
			expected: false,
		},
		{
			name:     "Invalid date format - wrong time format",
			dateStr:  "2025-06-23 22:49",
			expected: false,
		},
		{
			name:     "Empty string",
			dateStr:  "",
			expected: false,
		},
		{
			name:     "Random string",
			dateStr:  "not-a-date",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateCorrect(tt.dateStr)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for date: %s", tt.expected, result, tt.dateStr)
			}
		})
	}
}

func TestCreateDateFilter(t *testing.T) {
	tests := []struct {
		name          string
		initialQuery  string
		initialArgs   []any
		param         string
		operator      string
		expectedQuery string
		expectedArgs  []any
		expectedError bool
	}{
		{
			name:          "Valid date param",
			initialQuery:  "SELECT * FROM table WHERE 1=1",
			initialArgs:   []any{},
			param:         "2025-06-23 22:49:00",
			operator:      "created_at >=",
			expectedQuery: "SELECT * FROM table WHERE 1=1 AND created_at >= ?",
			expectedArgs:  []any{"2025-06-23 22:49:00"},
			expectedError: false,
		},
		{
			name:          "Empty param",
			initialQuery:  "SELECT * FROM table WHERE 1=1",
			initialArgs:   []any{},
			param:         "",
			operator:      "created_at >=",
			expectedQuery: "SELECT * FROM table WHERE 1=1",
			expectedArgs:  []any{},
			expectedError: false,
		},
		{
			name:          "Invalid date format",
			initialQuery:  "SELECT * FROM table WHERE 1=1",
			initialArgs:   []any{},
			param:         "2025-06-23",
			operator:      "created_at >=",
			expectedQuery: "SELECT * FROM table WHERE 1=1",
			expectedArgs:  []any{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.initialQuery
			args := tt.initialArgs

			err := createDateFilter(&query, &args, tt.param, tt.operator)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if query != tt.expectedQuery {
				t.Errorf("Expected query '%s', got '%s'", tt.expectedQuery, query)
			}

			if !reflect.DeepEqual(args, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, args)
			}
		})
	}
}

func TestApplyDateFilters(t *testing.T) {
	tests := []struct {
		name             string
		initialQuery     string
		initialArgs      []any
		paramOperatorMap map[string]string
		expectedQuery    string
		expectedArgs     []any
		expectedError    bool
	}{
		{
			name:         "Multiple valid date filters",
			initialQuery: "SELECT * FROM table WHERE 1=1",
			initialArgs:  []any{},
			paramOperatorMap: map[string]string{
				"2025-06-23 22:49:00": "created_at >=",
				"2025-06-24 22:49:00": "created_at <=",
			},
			expectedQuery: "SELECT * FROM table WHERE 1=1 AND created_at >= ? AND created_at <= ?",
			expectedArgs:  []any{"2025-06-23 22:49:00", "2025-06-24 22:49:00"},
			expectedError: false,
		},
		{
			name:             "Empty map",
			initialQuery:     "SELECT * FROM table WHERE 1=1",
			initialArgs:      []any{},
			paramOperatorMap: map[string]string{},
			expectedQuery:    "SELECT * FROM table WHERE 1=1",
			expectedArgs:     []any{},
			expectedError:    false,
		},
		{
			name:         "Invalid date in map",
			initialQuery: "SELECT * FROM table WHERE 1=1",
			initialArgs:  []any{},
			paramOperatorMap: map[string]string{
				"invalid-date": "created_at >=",
			},
			expectedQuery: "SELECT * FROM table WHERE 1=1",
			expectedArgs:  []any{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.initialQuery
			args := tt.initialArgs

			err := applyDateFilters(&query, &args, tt.paramOperatorMap)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if query != tt.expectedQuery {
				t.Errorf("Expected query '%s', got '%s'", tt.expectedQuery, query)
			}

			if !reflect.DeepEqual(args, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, args)
			}
		})
	}
}

func TestApplyLikeFilter(t *testing.T) {
	tests := []struct {
		name          string
		initialQuery  string
		initialArgs   []any
		filter        string
		columnName    string
		expectedQuery string
		expectedArgs  []any
	}{
		{
			name:          "Non-empty filter",
			initialQuery:  "SELECT * FROM table WHERE 1=1",
			initialArgs:   []any{},
			filter:        "test",
			columnName:    "name",
			expectedQuery: "SELECT * FROM table WHERE 1=1 AND name LIKE ?",
			expectedArgs:  []any{"%" + "test%"},
		},
		{
			name:          "Empty filter",
			initialQuery:  "SELECT * FROM table WHERE 1=1",
			initialArgs:   []any{},
			filter:        "",
			columnName:    "name",
			expectedQuery: "SELECT * FROM table WHERE 1=1",
			expectedArgs:  []any{},
		},
		{
			name:          "Filter with existing args",
			initialQuery:  "SELECT * FROM table WHERE id = ?",
			initialArgs:   []any{123},
			filter:        "john",
			columnName:    "username",
			expectedQuery: "SELECT * FROM table WHERE id = ? AND username LIKE ?",
			expectedArgs:  []any{123, "%john%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.initialQuery
			args := tt.initialArgs

			applyLikeFilter(&query, &args, tt.filter, tt.columnName)

			if query != tt.expectedQuery {
				t.Errorf("Expected query '%s', got '%s'", tt.expectedQuery, query)
			}

			if !reflect.DeepEqual(args, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, args)
			}
		})
	}
}

func TestApplyPagination(t *testing.T) {
	tests := []struct {
		name          string
		initialQuery  string
		initialArgs   []any
		limitStr      string
		offsetStr     string
		expectedQuery string
		expectedArgs  []any
	}{
		{
			name:          "Valid limit and offset",
			initialQuery:  "SELECT * FROM table",
			initialArgs:   []any{},
			limitStr:      "10",
			offsetStr:     "5",
			expectedQuery: "SELECT * FROM table LIMIT ? OFFSET ?",
			expectedArgs:  []any{10, 5},
		},
		{
			name:          "Empty strings use defaults",
			initialQuery:  "SELECT * FROM table",
			initialArgs:   []any{},
			limitStr:      "",
			offsetStr:     "",
			expectedQuery: "SELECT * FROM table LIMIT ? OFFSET ?",
			expectedArgs:  []any{20, 0},
		},
		{
			name:          "Invalid strings use defaults",
			initialQuery:  "SELECT * FROM table",
			initialArgs:   []any{},
			limitStr:      "abc",
			offsetStr:     "xyz",
			expectedQuery: "SELECT * FROM table LIMIT ? OFFSET ?",
			expectedArgs:  []any{20, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.initialQuery
			args := tt.initialArgs

			applyPagination(&query, &args, tt.limitStr, tt.offsetStr)

			if query != tt.expectedQuery {
				t.Errorf("Expected query '%s', got '%s'", tt.expectedQuery, query)
			}

			if !reflect.DeepEqual(args, tt.expectedArgs) {
				t.Errorf("Expected args %v, got %v", tt.expectedArgs, args)
			}
		})
	}
}

func TestApplySorting(t *testing.T) {
	tests := []struct {
		name           string
		initialQuery   string
		sortBy         string
		sortOrder      string
		allowedColumns []string
		defaultColumn  string
		expectedQuery  string
	}{
		{
			name:           "Valid sort column and order",
			initialQuery:   "SELECT * FROM table",
			sortBy:         "name",
			sortOrder:      "ASC",
			allowedColumns: []string{"name", "created_at", "id"},
			defaultColumn:  "created_at",
			expectedQuery:  "SELECT * FROM table ORDER BY name ASC",
		},
		{
			name:           "Invalid sort column uses default",
			initialQuery:   "SELECT * FROM table",
			sortBy:         "invalid_column",
			sortOrder:      "ASC",
			allowedColumns: []string{"name", "created_at", "id"},
			defaultColumn:  "created_at",
			expectedQuery:  "SELECT * FROM table ORDER BY created_at ASC",
		},
		{
			name:           "Invalid sort order uses DESC",
			initialQuery:   "SELECT * FROM table",
			sortBy:         "name",
			sortOrder:      "INVALID",
			allowedColumns: []string{"name", "created_at", "id"},
			defaultColumn:  "created_at",
			expectedQuery:  "SELECT * FROM table ORDER BY name DESC",
		},
		{
			name:           "Empty sort order uses DESC",
			initialQuery:   "SELECT * FROM table",
			sortBy:         "name",
			sortOrder:      "",
			allowedColumns: []string{"name", "created_at", "id"},
			defaultColumn:  "created_at",
			expectedQuery:  "SELECT * FROM table ORDER BY name DESC",
		},
		{
			name:           "DESC order",
			initialQuery:   "SELECT * FROM table",
			sortBy:         "created_at",
			sortOrder:      "DESC",
			allowedColumns: []string{"name", "created_at", "id"},
			defaultColumn:  "created_at",
			expectedQuery:  "SELECT * FROM table ORDER BY created_at DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.initialQuery

			applySorting(&query, tt.sortBy, tt.sortOrder, tt.allowedColumns, tt.defaultColumn)

			if query != tt.expectedQuery {
				t.Errorf("Expected query '%s', got '%s'", tt.expectedQuery, query)
			}
		})
	}
}
