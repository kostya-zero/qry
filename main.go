package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

type DriverConfig struct {
	Name    string
	SQLName string
	Dialect string
}

var supportedDrivers = map[string]DriverConfig{
	"postgres": {
		Name:    "postgres",
		SQLName: "pgx",
		Dialect: "postgres",
	},
	"sqlite": {
		Name:    "sqlite",
		SQLName: "sqlite",
		Dialect: "sqlite",
	},
}

func detectDriver(dsn string) (DriverConfig, error) {
	lower := strings.ToLower(dsn)

	switch {
	case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
		return supportedDrivers["postgres"], nil
	case strings.HasPrefix(lower, "sqlite:"), strings.HasPrefix(lower, "file:"):
		return supportedDrivers["sqlite"], nil
	case strings.HasSuffix(lower, ".db"), strings.HasSuffix(lower, ".sqlite"), strings.HasSuffix(lower, ".sqlite3"), lower == ":memory:":
		return supportedDrivers["sqlite"], nil
	default:
		return DriverConfig{}, errors.New("cannot detect driver from DSN; specify it explicitly with --driver")
	}
}

func resolveDriver(driver, dsn string) (DriverConfig, error) {
	if driver == "" {
		return detectDriver(dsn)
	}

	if _, ok := supportedDrivers[driver]; !ok {
		return DriverConfig{}, errors.New("specified driver is not supported")
	}

	return supportedDrivers[driver], nil
}

func run(driver, dsn string) error {
	if dsn == "" {
		envDsn, ok := os.LookupEnv("DATABASE_URL")
		envDsn = strings.TrimSpace(envDsn)
		if !ok || envDsn == "" {
			return errors.New("DSN is required")
		} else {
			dsn = envDsn
		}
	}

	databaseDriver, err := resolveDriver(driver, dsn)
	if err != nil {
		return err
	}

	conn, err := sql.Open(databaseDriver.SQLName, dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		return err
	}

	dialect, ok := DialectRegistry[databaseDriver.Dialect]
	if !ok {
		return errors.New("unknown driver")
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	session := NewSession(dialect, conn, cfg)
	return session.RunREPL()
}

func main() {
	setupColors()

	var (
		driver      string
		dsn         string
		showDrivers bool
	)

	rootCmd := &cobra.Command{
		Use:           "qry [DSN]",
		Short:         "Universal SQL CLI query runner",
		Long:          "A CLI query runner with support for multiple databases with SQL-like syntax.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if showDrivers {
				for _, v := range supportedDrivers {
					fmt.Println(v.Name)
				}
				return nil
			}

			if len(args) == 1 {
				dsn = strings.TrimSpace(args[0])
			}

			if dsn == "" && os.Getenv("DATABASE_URL") == "" {
				return cmd.Help()
			}

			return run(driver, dsn)
		},
	}

	rootCmd.Flags().StringVarP(&driver, "driver", "d", "", "name of driver to use")
	rootCmd.Flags().BoolVar(&showDrivers, "list-drivers", false, "show all available drivers")

	if err := rootCmd.Execute(); err != nil {
		PrintError(err.Error())
		os.Exit(1)
	}
}
