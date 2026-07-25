use colored::Colorize;

pub fn print_error(msg: &str) {
    println!("{}: {}", "error".bright_red().bold(), msg);
}
