# Quro

**Quro** is a command-line interface (CLI) tool designed to provide an interactive query runner for various databases, including PostgreSQL and SQLite. It provides a REPL-like experience for executing SQL queries against connected databases.

## Features

- **Interactive REPL**: Execute SQL queries in real-time.
- **Multi-Database Support**: Currently supports PostgreSQL and SQLite.
- **Easy Configuration**: Connect using standard Data Source Names (DSN).

## Installation

If you have Rust toolchain installed, run this command:

```bash
cargo install quro
```

You can go to [GitHub Releases](https://github.com/kostya-zero/quro/releases) page and download binary for your OS and architecture.

## Usage

Run the tool by providing a DSN:

```bash
# Provide a SQLite DSN
quro file:my_database.db

# Provide a PostgreSQL DSN
quro "postgres://user:password@localhost:5432/dbname"
```

The driver is automatically detected from the DSN. You can also specify it explicitly or use an environment variable:

```bash
# Using DATABASE_URL environment variable
DATABASE_URL="postgres://user:password@localhost:5432/dbname" quro

# Specifying driver explicitly
quro -d postgres "user=myuser password=mypass dbname=mydb"
```

### Options

- `-d, --driver`: Specify the database driver (`postgres` or `sqlite`).
- `--list-drivers`: List all supported database drivers.

## License

This project is licensed under the terms of the MIT license.
