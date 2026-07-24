use clap::Parser;

/// A command-line query runner.
#[derive(Parser)]
#[command(
    name = "qry",
    about = env!("CARGO_PKG_DESCRIPTION"),
    version = env!("CARGO_PKG_VERSION"),
    arg_required_else_help = true

)]
pub struct Cli {
    /// A database URL that QRY had to connect to.
    pub database_url: String,

    /// A driver to use. Usually determined from database URL,
    /// but can be specified explicitly.
    #[arg(short, long)]
    pub driver: Option<String>,
}
