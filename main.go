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
	driver      string
	dsn         string
	showDrivers bool
)

var supportedDrivers = []string{"postgres", "sqlite"}

func mainloop() error {
	if showDrivers {
		for _, v := range supportedDrivers {
			fmt.Println(v)
		}
		return nil
	}

	envDsn, ok := os.LookupEnv("DATABASE_URL")
	if ok && dsn == "" {
		dsn = envDsn
	}

	if dsn == "" {
		PrintWarn("no DSN provided, using ':memory:' instead.")
		dsn = ":memory:"
	}

	var isSupportedDriver bool
	for _, v := range supportedDrivers {
		if driver == v {
			isSupportedDriver = true
		}
	}

	if !isSupportedDriver {
		return errors.New("driver is not supported")
	}

	if driver == "postgres" {
		driver = "pgx"
	}

	conn, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	var dialect Dialect
	switch driver {
	case "pgx":
		dialect = PostgresDialect{}
	case "sqlite":
		dialect = SQLiteDialect{}
	default:
		return errors.New("unknown driver")
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

	rootCmd.Flags().StringVarP(&driver, "driver", "d", "sqlite", "name of driver to use ")
	rootCmd.Flags().BoolVar(&showDrivers, "list-drivers", false, "show all available drivers")

	if err := rootCmd.Execute(); err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}
