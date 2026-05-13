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

func detectDriver(dsn string) string {
	dsn = strings.ToLower(dsn)

	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") || (strings.Contains(dsn, "host=") && strings.Contains(dsn, "user=")) {
		return "postgres"
	}

	if strings.HasPrefix(dsn, "sqlite:") || strings.HasPrefix(dsn, "file:") || strings.Contains(dsn, ".db") || strings.Contains(dsn, ".sqlite3") {
		return "sqlite"
	}

	return ""
}

func getDriver() error {
	if driver == "" {
		detectedDriver := detectDriver(dsn)
		if detectedDriver == "" {
			return errors.New("cannot detect driver from DSN")
		}
		driver = detectedDriver
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

	return nil
}

func getDSN() error {
	if dsn != "" {
		return nil
	}

	envDsn, ok := os.LookupEnv("DATABASE_URL")
	if ok && dsn == "" {
		dsn = envDsn
	}

	return nil
}

func mainloop() error {
	err := getDSN()
	if err != nil {
		return err
	}

	err = getDriver()
	if err != nil {
		return err
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
	setupColors()
	rootCmd := &cobra.Command{
		Use:           "qry [DSN]",
		Short:         "Query runner for PostgreSQL",
		Long:          "A CLI query runner with support for multiple databases with SQL-like syntax.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if showDrivers {
				for _, v := range supportedDrivers {
					fmt.Println(v)
				}
				return nil
			}

			if len(args) == 0 {
				cmd.Help()
				return nil
			}

			if len(args) == 1 {
				dsn = strings.TrimSpace(args[0])
			}

			err := mainloop()
			if err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&driver, "driver", "d", "", "name of driver to use ")
	rootCmd.Flags().BoolVar(&showDrivers, "list-drivers", false, "show all available drivers")

	if err := rootCmd.Execute(); err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}
