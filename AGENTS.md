# QRY

`qry` is a Go CLI that opens an interactive SQL REPL for PostgreSQL and SQLite. 
Input DSN can be passed as an argument or via `DATABASE_URL`.
Driver can be auto-detected from DSN or explicitly set with `-d/--driver`.

## Architecture
- `main.go`: CLI entrypoint (Cobra), argument/env handling, driver detection, DB connection bootstrap.
- `session.go`: interactive REPL loop, internal commands (`.tables`, `.schema`, `.db`, `.stats`, `.help`, `.exit`), query execution, table rendering.
- `dialect.go`: DB-specific dialect abstraction and SQL snippets for PostgreSQL/SQLite metadata.
- `styles.go`, `errors.go`, `terminal.go`, `version.go`: UI styling, error output, terminal behavior, version constants.

## Development workflow (just)
- Run app:
  - `just run`
- Build binary:
  - `just build`

## Agent notes
- Keep changes minimal and focused.
- Prefer existing project patterns (simple structs/functions, no over-abstraction).
- Validate changes with `go vet` before finalizing.
