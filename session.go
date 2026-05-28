package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/chzyer/readline"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	dialect      Dialect
	conn         *sqlx.DB
	successCount int
	totalQueries int
	startTime    time.Time
}

func NewSession(dialect Dialect, conn *sqlx.DB) *Session {
	return &Session{dialect: dialect, conn: conn, successCount: 0, startTime: time.Now()}
}

func (s *Session) handleInternalCommand(command string) error {
	splitted := strings.SplitN(command, " ", 2)
	cmd := splitted[0]
	switch cmd {
	case ".tables":
		_, rows, err := s.ExecuteQuery(s.dialect.GetTablesQuery())
		if err != nil {
			PrintError("database error: " + err.Error())
			return nil
		}
		for _, v := range rows {
			fmt.Println(v[0])
		}
	case ".db":
		cols, rows, err := s.ExecuteQuery(s.dialect.GetDatabasesQuery())
		if err != nil {
			PrintError("database error: " + err.Error())
			return nil
		}
		s.renderTable(cols, rows)
	case ".schema":
		if len(splitted) == 1 {
			PrintError("Table name is required.")
			return nil
		}

		tableName := strings.TrimSpace(splitted[1])
		cols, rows, err := s.ExecuteQuery(s.dialect.GetTableSchema(tableName))
		if err != nil {
			PrintError("database error: " + err.Error())
			return nil
		}
		s.renderTable(cols, rows)
	case ".stats":
		s.PrintStats()
	case ".exit":
		s.PrintStats()
		fmt.Println(lipgloss.NewStyle().Foreground(ColorPrimary).Render("Goodbye!"))
		return ErrExit
	case ".version":
		fmt.Println(WelcomeStyle.Render(fmt.Sprintf("QRY Shell v%s", QryVersion)))
	case ".help":
		rows := [][]string{
			{".tables", "display all tables in current database"},
			{".schema <table>", "display schema of the table"},
			{".version", "display version of QRY"},
			{".stats", "display stats for current session"},
			{".db", "print information about databases"},
			{".help", "shows this help message"},
			{".exit", "close QRY"},
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(TableBorderStyle).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if col == 0 {
					return InternalCmdStyle.Padding(0, 1)
				}
				return DescStyle.Padding(0, 1)
			})
		fmt.Println(t.Render())
	default:
		PrintError("Unknown internal command: " + command)
	}

	return nil
}

func (s *Session) PrintStats() {
	rows := [][]string{
		{"Session Time", time.Since(s.startTime).Round(time.Second).String()},
		{"Queries Stats", fmt.Sprintf("%d success, %d error", s.successCount, s.totalQueries-s.successCount)},
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(TableBorderStyle).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return TableHeaderStyle.Align(lipgloss.Left)
			}
			return TableCellStyle
		})

	fmt.Println(t.Render())
}

func (s *Session) handleSQLQuery(query string) error {
	query = s.sanitizeQuery(query)
	if query == "" {
		return nil
	}
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
		values, err := rows.SliceScan()
		if err != nil {
			return nil, nil, err
		}

		var row []string
		for _, val := range values {
			row = append(row, s.formatValue(val))
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
		Border(lipgloss.RoundedBorder()).
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

func (s *Session) RunREPL() error {
	config := &readline.Config{
		Prompt:          "qry>",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}
	rl, err := readline.NewEx(config)
	if err != nil {
		return ErrReadLineFail
	}
	defer rl.Close()

	fmt.Println(WelcomeStyle.Render(fmt.Sprintf("QRY Shell v%s", QryVersion)))
	fmt.Println(SubtextStyle.Render("Use '.help' for commands."))
	fmt.Println()

	var buffer string
	for {
		prompt := "qry> "
		if buffer != "" {
			prompt = "..> "
		}

		rl.SetPrompt(prompt)
		line, err := rl.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				buffer = ""
				continue
			}

			if errors.Is(err, io.EOF) {
				s.PrintStats()
				fmt.Println(lipgloss.NewStyle().Foreground(ColorPrimary).Render("Goodbye!"))
				return nil
			}

			return fmt.Errorf("readline error: %w", err)
		}

		cleanLine := strings.TrimSpace(string(line))
		if cleanLine == "" {
			continue
		}

		if buffer == "" && strings.HasPrefix(cleanLine, ".") {
			err := s.handleInternalCommand(cleanLine)
			if err != nil {
				if errors.Is(err, ErrExit) {
					return nil
				}
			}
			continue
		}

		buffer += " " + cleanLine
		if strings.HasSuffix(cleanLine, ";") {
			s.totalQueries += 1
			err := s.handleSQLQuery(buffer)
			if err != nil {
				PrintError(fmt.Sprintf("database error occurred: %v", err))
			} else {
				s.successCount += 1
			}
			buffer = ""
		}
	}
}
