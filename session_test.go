package main

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestExecuteQuery(t *testing.T) {
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Exec("CREATE TABLE users (id INTEGER, name TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("INSERT INTO users VALUES (1, 'Ada')"); err != nil {
		t.Fatal(err)
	}

	session := NewSession(SQLiteDialect{}, conn, &Config{})
	columns, rows, err := session.ExecuteQuery("SELECT id, name FROM users")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(columns, []string{"id", "name"}) {
		t.Fatalf("columns = %v", columns)
	}
	if !reflect.DeepEqual(rows, [][]string{{"1", "Ada"}}) {
		t.Fatalf("rows = %v", rows)
	}
}
