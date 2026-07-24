use anyhow::Result;
use std::str::FromStr;

pub mod sqlite;

#[derive(Debug)]
pub enum Drivers {
    Sqlite,
}

impl FromStr for Drivers {
    type Err = String;
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_ascii_lowercase().as_str() {
            "sqlite" => Ok(Self::Sqlite),
            _ => Err("unknown driver".to_string()),
        }
    }
}

#[derive(Debug, Clone, Default)]
pub struct QueryOutput {
    pub columns: Vec<String>,
    pub rows: Vec<Vec<String>>,
    pub affected_rows: usize,
}

pub trait Driver {
    fn get_tables_query() -> &'static str;
    fn get_databases_query() -> &'static str;
    fn get_tables_schema(table: &str) -> String;
    fn name() -> &'static str;
    fn execute_query(&self, query: &str) -> Result<QueryOutput>;
}
