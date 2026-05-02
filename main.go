package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	provider string
	dsn      string
)

func mainloop() error {
	if provider == "postgres" {
		provider = "pgx"
	}

	conn, err := sqlx.Connect(provider, dsn)
	if err != nil {
		return err
	}

	var dialect Dialect
	switch provider {
	case "pgx":
		dialect = PostgresDialect{}
	default:
		return errors.New("unknown provider")
	}

	session := NewSession(dialect, conn)
	session.RunREPL()

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:           "qry [DSN]",
		Short:         "Query runner for PostgreSQL",
		Long:          "QRY is a query runner that can send queries to PostgreSQL in more fancy way.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(1)
			}

			if len(args) > 1 {
				fmt.Println("Too many arguments.")
				os.Exit(1)
			}

			dsn = strings.TrimSpace(args[0])

			err := mainloop()
			if err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&provider, "provider", "p", "postgres", "which provider to use (postgres, sqlite)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("an error occured: %v", err)
		os.Exit(1)
	}
}
