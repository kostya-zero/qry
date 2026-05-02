package main

type Dialect interface {
	GetTablesQuery() string
}

type PostgresDialect struct{}

func (p PostgresDialect) GetTablesQuery() string {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
}
