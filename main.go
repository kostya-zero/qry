package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var (
	provider      string
	dsn           string
	showProviders bool
)

var supportedProviders = []string{"postgres", "sqlite"}

func mainloop() error {
	if showProviders {
		for _, v := range supportedProviders {
			fmt.Println(v)
		}
		return nil
	}

	if dsn == "" {
		PrintWarn("no DSN provided, using ':memory:' instead.")
		dsn = ":memory:"
	}

	var isSupportedProvider bool
	for _, v := range supportedProviders {
		if provider == v {
			isSupportedProvider = true
		}
	}

	if !isSupportedProvider {
		return errors.New("provider is not supported")
	}

	if provider == "postgres" {
		provider = "pgx"
	}

	conn, err := sqlx.Connect(provider, dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	var dialect Dialect
	switch provider {
	case "pgx":
		dialect = PostgresDialect{}
	case "sqlite":
		dialect = SQLiteDialect{}
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
		Long:          "A CLI query runner with support for multiple databases with SQL-like syntax.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				dsn = strings.TrimSpace(args[0])
			}

			if len(args) > 1 {
				fmt.Println("Too many arguments.")
				os.Exit(1)
			}

			err := mainloop()
			if err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&provider, "provider", "p", "sqlite", "name of provider to use ")
	rootCmd.Flags().BoolVar(&showProviders, "list-providers", false, "show all available providers")

	if err := rootCmd.Execute(); err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}
