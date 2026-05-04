package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/jmoiron/sqlx"
	"github.com/smartystreets/cle"
)

type CommandInfo struct {
	Usage       string
	Description string
}

type Session struct {
	dialect Dialect
	limit   int
	conn    *sqlx.DB
}

func NewSession(dialect Dialect, conn *sqlx.DB) *Session {
	return &Session{dialect: dialect, limit: 100, conn: conn}
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
	case ".limit":
		if len(splitted) < 2 {
			fmt.Printf("Current limit is %d.\n", s.limit)
			return
		}

		newLimit, err := strconv.Atoi(splitted[1])
		if err != nil {
			PrintError("wrong value")
			return
		}
		if newLimit <= 0 {
			PrintError("value should be greter than zero")
			return
		}
		if newLimit > 600 {
			PrintWarn("new limit is very high, it could lead to memory leak")
		}

		s.limit = newLimit
		fmt.Printf("New limit is set to %d.\n", newLimit)
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
	case ".version":
		fmt.Printf("QRY Shell v%s\n", QryVersion)
	case ".help":
		commands := []CommandInfo{
			{Usage: ".tables", Description: "display all tables in current database"},
			{Usage: ".schema <table>", Description: "display schema of the table"},
			{Usage: ".limit <value>", Description: "display or set limit"},
			{Usage: ".exit", Description: "close QRY"},
			{Usage: ".version", Description: "display version of QRY"},
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
					return lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
				}
				return lipgloss.NewStyle()
			})
		fmt.Println(t.Render())
	default:
		fmt.Printf("Unknown internal command: %s\n", command)
	}
}

func (s *Session) handleSQLQuery(query string) error {
	query = s.sanitizeQuery(query)
	headers, data, err := s.ExecuteQuery(query)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	s.renderTable(headers, data)
	return nil
}

func (s *Session) sanitizeQuery(query string) string {
	query = strings.TrimSpace(query)
	query = strings.TrimSuffix(query, ";")

	// Add LIMIT to not overflow the memory
	upperQuery := strings.ToUpper(query)
	if strings.HasPrefix(upperQuery, "SELECT") && !strings.Contains(upperQuery, "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", query, s.limit)
	}

	return s.conn.Rebind(query)
}

func (s *Session) formatValue(val any) string {
	if val == nil {
		return "NULL"
	}
	if b, ok := val.([]byte); ok {
		var finalData string
		if !utf8.Valid(b) {
			finalData = hex.EncodeToString(b)
		} else {
			finalData = string(b)
		}
		return finalData
	} else {
		return fmt.Sprintf("%v", val)
	}
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
		if err := rows.MapScan(rowMap); err != nil {
			return nil, nil, err
		}

		var row []string
		for _, col := range cols {
			row = append(row, s.formatValue(rowMap[col]))
		}
		tableData = append(tableData, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return cols, tableData, nil
}

func (s *Session) renderTable(headers []string, data [][]string) {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(TableBorderStyle).
		Headers(headers...).
		Rows(data...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return TableHeaderStyle
			default:
				return TableCellStyle
			}
		})

	fmt.Println(t.Render())
}

func (s *Session) RunREPL() {
	lineEditor := cle.NewCLE()

	fmt.Printf("QRY Shell v%s\n", QryVersion)
	fmt.Println("Use '.help' for commands.")

	var buffer string
	for {
		prompt := "qry> "
		if buffer != "" {
			prompt = "..> "
		}

		promptString := PromptStyle.Render(prompt)
		line := lineEditor.ReadInput(promptString)

		cleanLine := strings.TrimSpace(string(line))
		if cleanLine == "" {
			continue
		}

		if buffer == "" && strings.HasPrefix(cleanLine, ".") {
			s.handleInternalCommand(cleanLine)
			continue
		}

		buffer += " " + cleanLine
		if strings.HasSuffix(cleanLine, ";") {
			err := s.handleSQLQuery(buffer)
			if err != nil {
				PrintError(fmt.Sprintf("database error occurred: %v", err))
			}
			buffer = ""
		}
	}
}
