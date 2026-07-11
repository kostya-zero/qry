package main

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/ergochat/readline"
)

type Session struct {
	dialect      Dialect
	conn         *sql.DB
	cfg          *Config
	border       lipgloss.Border
	successCount int
	totalQueries int
	startTime    time.Time
}

func NewSession(dialect Dialect, conn *sql.DB, cfg *Config) *Session {
	var border lipgloss.Border
	switch cfg.Borders {
	case "none":
		border = lipgloss.Border{}
	case "rounded":
		border = lipgloss.RoundedBorder()
	case "normal":
		border = lipgloss.NormalBorder()
	case "ascii":
		border = lipgloss.ASCIIBorder()
	case "block":
		border = lipgloss.BlockBorder()
	case "double":
		border = lipgloss.DoubleBorder()
	case "hidden":
		border = lipgloss.HiddenBorder()
	case "thick":
		border = lipgloss.ThickBorder()
	default:
		border = lipgloss.RoundedBorder()
	}

	return &Session{dialect: dialect, conn: conn, successCount: 0, startTime: time.Now(), cfg: cfg, border: border}
}

func (s *Session) handleInternalCommand(command string) error {
	splitted := strings.SplitN(command, " ", 2)
	cmd := splitted[0]
	switch cmd {
	case ".tables":
		_, rows, err := s.ExecuteQuery(s.dialect.GetTablesQuery())
		if err != nil {
			return fmt.Errorf("databaser error: %w", err)
		}
		for _, row := range rows {
			if len(row) > 0 {
				fmt.Println(row[0])
			}
		}
	case ".db":
		cols, rows, err := s.ExecuteQuery(s.dialect.GetDatabasesQuery())
		if err != nil {
			return fmt.Errorf("databaser error: %w", err)
		}
		s.renderTable(cols, rows)
	case ".schema":
		if len(splitted) == 1 {
			return errors.New("table name is required")
		}

		tableName := strings.TrimSpace(splitted[1])
		cols, rows, err := s.ExecuteQuery(s.dialect.GetTableSchema(tableName))
		if err != nil {
			return fmt.Errorf("databaser error: %w", err)
		}
		s.renderTable(cols, rows)
	case ".stats":
		s.PrintStats()
	case ".config":
		rows := [][]string{
			{"borders", s.cfg.Borders},
		}

		t := table.New().
			Border(s.border).
			BorderStyle(TableBorderStyle).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if col == 0 {
					return InternalCmdStyle.Padding(0, 1)
				}
				return DescStyle.Padding(0, 1)
			})
		fmt.Println(t.Render())
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
			{".config", "print current configuration"},
			{".help", "shows this help message"},
			{".exit", "close QRY"},
		}

		t := table.New().
			Border(s.border).
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
		return fmt.Errorf("unknown internal command: %s", command)
	}

	return nil
}

func (s *Session) PrintStats() {
	rows := [][]string{
		{"Session Time", time.Since(s.startTime).Round(time.Second).String()},
		{"Queries Stats", fmt.Sprintf("%d success, %d error", s.successCount, s.totalQueries-s.successCount)},
	}

	t := table.New().
		Border(s.border).
		BorderStyle(TableBorderStyle).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return InternalCmdStyle.Padding(0, 1)
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

	if s.isQueryReturningRows(query) {
		headers, data, err := s.ExecuteQuery(query)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			fmt.Println(SubtextStyle.Render("No rows."))
			return nil
		}

		s.renderTable(headers, data)
		return nil
	}

	result, err := s.conn.Exec(query)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil {
		fmt.Println(SubtextStyle.Render(fmt.Sprintf("OK, %d rows affected.", rowsAffected)))
	} else {
		fmt.Println(SubtextStyle.Render("OK."))
	}

	return nil
}

func (s *Session) sanitizeQuery(query string) string {
	query = strings.TrimSpace(query)
	query = strings.TrimSuffix(query, ";")
	return strings.TrimSpace(query)
}

func (s *Session) isQueryReturningRows(query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))

	return strings.HasPrefix(query, "select") ||
		strings.HasPrefix(query, "show") ||
		strings.HasPrefix(query, "pragma") ||
		strings.HasPrefix(query, "with") ||
		strings.HasPrefix(query, "explain") ||
		strings.HasPrefix(query, "describe")
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
	rows, err := s.conn.Query(query)
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
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}

		err := rows.Scan(pointers...)
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
		Border(s.border).
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
				PrintError(err.Error())
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
