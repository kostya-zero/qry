package main

import "fmt"

var DialectRegistry = map[string]Dialect{
	"postgres": PostgresDialect{},
	"sqlite":   SQLiteDialect{},
}

type Dialect interface {
	GetTablesQuery() string
	GetTableSchema(table string) string
	GetDatabasesQuery() string
}

type PostgresDialect struct{}

func (p PostgresDialect) GetTablesQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
}

func (p PostgresDialect) GetTableSchema(table string) string {
	return fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", table)
}

func (p PostgresDialect) GetDatabasesQuery() string {
	return `
		SELECT 
            datname AS "name",
            pg_get_userbyid(datdba) AS "owner",
            pg_size_pretty(pg_database_size(datname)) AS "size",
            pg_encoding_to_char(encoding) AS "encoding",
            datcollate AS "collate"
        FROM pg_database
        WHERE datistemplate = false
        ORDER BY pg_database_size(datname) DESC
	`
}

type SQLiteDialect struct{}

func (s SQLiteDialect) GetTablesQuery() string {
	return "SELECT name FROM sqlite_master WHERE type='table'"
}

func (s SQLiteDialect) GetTableSchema(table string) string {
	return fmt.Sprintf("PRAGMA table_info('%s')", table)
}

func (s SQLiteDialect) GetDatabasesQuery() string {
	return "PRAGMA database_list"
}
