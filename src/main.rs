use std::{env, process::exit, str::FromStr};

use clap::{CommandFactory, Parser};

use crate::{
    cli::Cli,
    drivers::{Drivers, postgres::PostgresDriver, sqlite::SqliteDriver},
    session::Session,
    terminal::print_error,
};

mod cli;
mod drivers;
mod session;
mod terminal;

fn detect_driver(dsn: &str) -> Option<Drivers> {
    let lower_string = dsn.to_ascii_lowercase();
    let lower = lower_string.as_str();

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

fn main() {
    let args = Cli::parse();
    if args.list_drivers {
        println!("postgres\nsqlite");
        return;
    }

    let mut cmd = Cli::command();

    let database_url = if let Some(u) = args.database_url {
        u
    } else {
        let database_url = env::var("DATABASE_URL");
        if database_url.is_err() {
            cmd.print_help().unwrap();
            exit(1)
        }
        database_url.unwrap()
    };

    let driver_to_use = if let Some(d) = args.driver {
        match Drivers::from_str(&d) {
            Ok(dr) => dr,
            Err(e) => {
                print_error(&e.to_string());
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

    match driver_to_use {
        Drivers::Sqlite => {
            let driver = SqliteDriver::new(&database_url);
            match driver {
                Ok(d) => {
                    let mut session = Session::new(d);
                    session.run_repl().unwrap();
                }
                Err(e) => {
                    print_error(&format!("Failed to connect to the database: {e}"));
                    exit(1)
                }
            }
        }
        Drivers::Postgres => {
            let driver = PostgresDriver::new(&database_url);
            match driver {
                Ok(d) => {
                    let mut session = Session::new(d);
                    session.run_repl().unwrap();
                }
                Err(e) => {
                    print_error(&format!("Failed to connect to the database: {e}"));
                    exit(1)
                }
            }
        }
    }
}
