package main

import "fmt"

type Dialect interface {
	GetTablesQuery() string
	GetTableSchema(table string) string
}

type PostgresDialect struct{}

func (p PostgresDialect) GetTablesQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
}

func (p PostgresDialect) GetTableSchema(table string) string {
	return fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", table)
}
