package database

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var tables = []interface{}{
	// models.User{},
}

func RunMigrations(db *sqlx.DB) error {
	for _, table := range tables {
		if err := createTableIfNotExists(db, table); err != nil {
			return err
		}
	}
	return nil
}

func createTableIfNotExists(db *sqlx.DB, model interface{}) error {
	tableName := getTableName(model)
	columns := getColumns(model)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			%s
		)
	`, tableName, strings.Join(columns, ",\n"))

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating table %s: %w", tableName, err)
	}

	return nil
}

func getTableName(model interface{}) string {
	t := reflect.TypeOf(model)
	return strings.ToLower(t.Name()) + "s"
}

func getColumns(model interface{}) []string {
	var columns []string
	t := reflect.TypeOf(model)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}

		column := fmt.Sprintf("%s %s", dbTag, getSQLType(field.Type, dbTag))
		columns = append(columns, column)
	}

	return columns
}

func getSQLType(t reflect.Type, dbTag string) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		if dbTag == "id" {
			return "SERIAL PRIMARY KEY"
		}
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.String:
		return "TEXT"
	case reflect.Bool:
		return "BOOLEAN"
	}

	if t == reflect.TypeOf(time.Time{}) {
		return "TIMESTAMP"
	}

	return "TEXT"
}
