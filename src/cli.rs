use clap::{ArgAction, Parser};

/// A command-line query runner.
#[derive(Parser)]
#[command(
    name = "quro",
    about = env!("CARGO_PKG_DESCRIPTION"),
    version = env!("CARGO_PKG_VERSION"),

)]
pub struct Cli {
    /// A database URL that QRY had to connect to.
    pub database_url: Option<String>,

    /// A driver to use. Usually determined from database URL,
    /// but can be specified explicitly.
    #[arg(short, long)]
    pub driver: Option<String>,

    /// Execute a specified query instead of launching REPL.
    #[arg(short, long)]
    pub query: Option<String>,

    /// Print a list of all supported drivers.
    #[arg(short, long, action = ArgAction::SetTrue)]
    pub list_drivers: bool,
}
