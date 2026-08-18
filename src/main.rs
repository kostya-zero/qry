use std::{env, process::exit};

use anyhow::Result;
use clap::{CommandFactory, Parser};

use crate::{
    cli::Cli,
    drivers::{Driver, Drivers, postgres::PostgresDriver, sqlite::SqliteDriver},
    session::Session,
    terminal::print_error,
};

mod cli;
mod drivers;
mod session;
mod terminal;

fn detect_driver(dsn: &str) -> Option<Drivers> {
    let lower = dsn.to_ascii_lowercase();

    if lower.starts_with("postgres://") || lower.starts_with("postgresql://") {
        return Some(Drivers::Postgres);
    }

    if lower.starts_with("sqlite:") || lower.starts_with("file:") {
        return Some(Drivers::Sqlite);
    }

    if lower.ends_with(".db")
        || lower.ends_with(".sqlite")
        || lower.ends_with(".sqlite3")
        || lower == ":memory:"
    {
        return Some(Drivers::Sqlite);
    }

    None
}

fn run_session<D: Driver>(driver: anyhow::Result<D>) {
    match driver {
        Ok(driver) => Session::new(driver).run_repl().unwrap(),
        Err(error) => {
            print_error(&format!("Failed to connect to the database: {error}"));
            exit(1)
        }
    }
}

fn execute_query<D: Driver>(driver: Result<D>, query: &str) {
    match driver {
        Ok(driver) => Session::new(driver).execute_query(query),
        Err(error) => {
            print_error(&format!("Failed to connect to the database: {error}"));
            exit(1)
        }
    }
}
fn main() {
    let args = Cli::parse();
    if args.list_drivers {
        println!("postgres\nsqlite");
        return;
    }

    let database_url = match args.database_url.or_else(|| env::var("DATABASE_URL").ok()) {
        Some(database_url) => database_url,
        None => {
            Cli::command().print_help().unwrap();
            exit(1)
        }
    };

    let driver_to_use = if let Some(d) = args.driver {
        match d.parse::<Drivers>() {
            Ok(dr) => dr,
            Err(e) => {
                print_error(&e);
                exit(1)
            }
        }
    } else if let Some(d) = detect_driver(&database_url) {
        d
    } else {
        print_error(
            "Failed to auto-detect database driver. Please, specify the driver name explicitly with '--driver'.",
        );
        exit(1)
    };

    if let Some(q) = args.query {
        match driver_to_use {
            Drivers::Sqlite => execute_query(SqliteDriver::new(&database_url), &q),
            Drivers::Postgres => execute_query(PostgresDriver::new(&database_url), &q),
        }
        return;
    }

    match driver_to_use {
        Drivers::Sqlite => run_session(SqliteDriver::new(&database_url)),
        Drivers::Postgres => run_session(PostgresDriver::new(&database_url)),
    }
}
