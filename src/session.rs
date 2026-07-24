use std::str::FromStr;

use anyhow::Result;
use rustyline::{DefaultEditor, error::ReadlineError};
use tabled::{builder::Builder, settings::Style};
use thiserror::Error;

use crate::drivers::{Driver, QueryOutput};

pub struct Session<D> {
    driver: D,
}

#[derive(Debug)]
pub enum Command {
    Version,
    Tables,
    Db,
    Schema,
    Help,
    Driver,
    Exit,
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum SessionError {
    #[error("unknown command: {0}")]
    CommandNotFound(String),

    #[error("{0}")]
    CommandError(String),

    #[error("exit")]
    Exit,
}

impl FromStr for Command {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_ascii_lowercase().as_str() {
            ".version" => Ok(Self::Version),
            ".db" => Ok(Self::Db),
            ".driver" => Ok(Self::Driver),
            ".tables" => Ok(Self::Tables),
            ".schema" => Ok(Self::Schema),
            ".help" => Ok(Self::Help),
            ".exit" | ".quit" => Ok(Self::Exit),
            _ => Err(format!("unknown command: {s}")),
        }
    }
}

impl<D> Session<D>
where
    D: Driver,
{
    pub fn new(driver: D) -> Self {
        Self { driver }
    }

    fn execute_internal_command(&mut self, command: &str) -> Result<(), SessionError> {
        let splitted: Vec<String> = command.splitn(2, " ").map(String::from).collect();

        let cmd = Command::from_str(&splitted[0])
            .map_err(|_| SessionError::CommandNotFound(command.to_string()))?;
        match cmd {
            Command::Version => println!("{}", env!("CARGO_PKG_VERSION")),
            Command::Exit => return Err(SessionError::Exit),
            Command::Tables => self.execute_query(D::get_tables_query()),
            Command::Db => self.execute_query(D::get_databases_query()),
            Command::Driver => println!("{}", D::name()),
            Command::Schema => {
                if splitted.len() < 2 {
                    return Err(SessionError::CommandError(
                        "table name is required".to_string(),
                    ));
                }

                let args = &splitted[1];
                if args.is_empty() {
                    return Err(SessionError::CommandError(
                        "table name is required".to_string(),
                    ));
                }

                self.execute_query(&D::get_tables_schema(args));
            }
            Command::Help => {
                let columns = vec!["command".to_string(), "description".to_string()];

                let rows = vec![
                    vec![".help".to_string(), "prints help message".to_string()],
                    vec![".version".to_string(), "prints version of qry".to_string()],
                    vec![".driver".to_string(), "prints driver name".to_string()],
                    vec![".exit, .quit".to_string(), "exit qry".to_string()],
                    vec![".db".to_string(), "prints all databases".to_string()],
                    vec![".tables".to_string(), "prints all tables".to_string()],
                    vec![
                        ".schema <table>".to_string(),
                        "prints schema of specific table".to_string(),
                    ],
                ];

                self.render_table(QueryOutput {
                    columns,
                    rows,
                    affected_rows: 0,
                });
            }
        }

        Ok(())
    }

    pub fn render_table(&self, data: QueryOutput) {
        let mut b = Builder::new();
        b.push_record(data.columns);

        for row in data.rows {
            b.push_record(row);
        }

        let mut t = b.build();
        t.with(Style::rounded());

        println!("{t}");
    }

    pub fn execute_query(&mut self, query: &str) {
        match self.driver.execute_query(query) {
            Ok(d) => {
                if d.columns.is_empty() && d.rows.is_empty() {
                    println!("OK, rows affected {}.", d.affected_rows);
                } else {
                    self.render_table(d);
                }
            }
            Err(e) => println!("database error: {e}"),
        }
    }

    pub fn run_repl(&mut self) -> Result<()> {
        let mut rl = DefaultEditor::new()?;

        println!("QRY v{}", env!("CARGO_PKG_VERSION"));
        let mut buf = String::new();
        loop {
            let mut prompt = String::new();

            if buf.is_empty() {
                prompt.push_str("qry> ");
            } else {
                prompt.push_str("...> ");
            }

            let readline = rl.readline(&prompt);
            match readline {
                Ok(line) => {
                    rl.add_history_entry(line.as_str())?;
                    let trimmed = line.trim();

                    if trimmed.is_empty() {
                        continue;
                    }

                    if trimmed.starts_with(".") {
                        let res = self.execute_internal_command(trimmed);
                        if let Err(e) = res {
                            if e == SessionError::Exit {
                                break;
                            } else {
                                eprintln!("{e}");
                            }
                        }
                        continue;
                    }

                    if !trimmed.ends_with(";") {
                        buf.push_str(&format!(" {}", trimmed));
                        continue;
                    }

                    buf.push_str(&format!(" {}", trimmed));
                    self.execute_query(&buf);
                    buf.clear();
                }
                Err(ReadlineError::Interrupted) => {
                    if !buf.is_empty() {
                        buf.clear();
                        println!("Buffer has been cleared.");
                        continue;
                    }
                }
                Err(ReadlineError::Eof) => {
                    break;
                }
                Err(err) => {
                    println!("Error: {:?}", err);
                    break;
                }
            }
        }

        println!("Goodbye!");

        Ok(())
    }
}
