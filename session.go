package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/jmoiron/sqlx"
	"github.com/peterh/liner"
)

type CommandInfo struct {
	Usage       string
	Description string
}

type Session struct {
	dialect Dialect
	conn    *sqlx.DB
}

func NewSession(dialect Dialect, conn *sqlx.DB) *Session {
	return &Session{dialect: dialect, conn: conn}
}

func (s *Session) handleInternalCommand(command string) {
	splitted := strings.SplitN(command, " ", 2)
	cmd := splitted[0]
	switch cmd {
	case ".tables":
		_, rows, err := s.ExecuteQuery(s.dialect.GetTablesQuery())
		if err != nil {
			fmt.Printf("database error: %v\n", err)
			return
		}
		for _, v := range rows {
			fmt.Println(v[0])
		}
	case ".schema":
		if len(splitted) == 1 {
			fmt.Println("Table name is required.")
			return
		}

		tableName := strings.TrimSpace(splitted[1])
		cols, rows, err := s.ExecuteQuery(s.dialect.GetTableSchema(tableName))
		if err != nil {
			fmt.Printf("database error: %v\n", err)
			return
		}
		s.renderTable(cols, rows)
	case ".exit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case ".help":
		commands := []CommandInfo{
			{Usage: ".tables", Description: "display all tables in current database"},
			{Usage: ".schema <table>", Description: "display schema of the table"},
			{Usage: ".exit", Description: "close QRY"},
			{Usage: ".help", Description: "shows this help message"},
		}

		var rows [][]string
		for _, c := range commands {
			row := []string{}
			row = append(row, c.Usage)
			row = append(row, c.Description)
			rows = append(rows, row)
		}

		t := table.New().
			Border(lipgloss.HiddenBorder()).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if col == 0 {
					return lipgloss.NewStyle().Bold(true)
				}
				return lipgloss.NewStyle()
			})
		fmt.Println(t.Render())
	default:
		fmt.Printf("Unknown internal command: %s\n", command)
	}
}

func (s *Session) handleSQLQuery(query string) {
	query = s.sanitizeQuery(query)
	headers, data, err := s.ExecuteQuery(query)
	if err != nil {
		fmt.Printf("database error: %v\n", err)
	}

	if len(data) == 0 {
		fmt.Println("No rows found.")
		return
	}

	s.renderTable(headers, data)
}

func (s *Session) sanitizeQuery(query string) string {
	query = strings.TrimSpace(query)
	query = strings.TrimSuffix(query, ";")

	// Add LIMIT to not overflow the memory
	upperQuery := strings.ToUpper(query)
	if strings.HasPrefix(upperQuery, "SELECT") && !strings.Contains(upperQuery, "LIMIT") {
		query = query + " LIMIT 100"
	}

	return s.conn.Rebind(query)
}

func (s *Session) ExecuteQuery(query string) ([]string, [][]string, error) {
	rows, err := s.conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var tableData [][]string
	for rows.Next() {
		rowMap := make(map[string]any)
		rows.MapScan(rowMap)

		var row []string
		for _, col := range cols {
			val := rowMap[col]
			if val == nil {
				row = append(row, "NULL")
			} else {
				if b, ok := val.([]byte); ok {
					row = append(row, string(b))
				} else {
					row = append(row, fmt.Sprintf("%v", val))
				}
			}
		}
		tableData = append(tableData, row)
	}

	return cols, tableData, nil
}

func (s *Session) renderTable(headers []string, data [][]string) {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // Розовый
		Bold(true).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1).Foreground(lipgloss.Color("251"))

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")) // Серый

	// Создаем таблицу
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		Rows(data...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return headerStyle
			default:
				return cellStyle
			}
		})

	fmt.Println(t.Render())
}

func (s *Session) RunREPL() {
	lineReader := liner.NewLiner()
	defer lineReader.Close()

	lineReader.SetCtrlCAborts(true)

	fmt.Println("QRY v0.1.0")
	fmt.Println("Use '.help' for commands.")

	var buffer string
	for {
		prompt := "qry> "
		if buffer != "" {
			prompt = "..> "
		}

		line, err := lineReader.Prompt(prompt)
		if err != nil {
			if err == liner.ErrPromptAborted {
				continue
			}
			break
		}

		cleanLine := strings.TrimSpace(line)
		if cleanLine == "" {
			continue
		}

		if buffer == "" && strings.HasPrefix(cleanLine, ".") {
			s.handleInternalCommand(cleanLine)
			lineReader.AppendHistory(cleanLine)
			continue
		}

		buffer += " " + cleanLine
		if strings.HasSuffix(cleanLine, ";") {
			s.handleSQLQuery(buffer)
			buffer = ""
			lineReader.AppendHistory(cleanLine)
		}
	}
}
