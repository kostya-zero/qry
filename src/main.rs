use std::process::exit;

use clap::Parser;

use crate::{
    cli::Cli,
    drivers::{Drivers, postgres::PostgresDriver, sqlite::SqliteDriver},
    session::Session,
};

mod cli;
mod drivers;
mod session;

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

    let driver_to_use = if let Some(d) = detect_driver(&args.database_url) {
        d
    } else {
        println!(
            "Failed to auto-detect database driver. Please, specify the driver name explicitly with '--driver'."
        );
        exit(1)
    };

    match driver_to_use {
        Drivers::Sqlite => {
            let driver = SqliteDriver::new(&args.database_url);
            match driver {
                Ok(d) => {
                    let mut session = Session::new(d);
                    session.run_repl().unwrap();
                }
                Err(e) => {
                    println!("Failed to connect to the database: {e}");
                    exit(1)
                }
            }
        }
        Drivers::Postgres => {
            let driver = PostgresDriver::new(&args.database_url);
            match driver {
                Ok(d) => {
                    let mut session = Session::new(d);
                    session.run_repl().unwrap();
                }
                Err(e) => {
                    println!("Failed to connect to the database: {e}");
                    exit(1)
                }
            }
        }
    }
}
