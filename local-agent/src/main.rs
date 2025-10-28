use clap::Parser;
use crossterm::style::Stylize;
use eyre::Result;
use std::process::ExitCode;

mod api;
mod cli;
mod util;

fn main() -> Result<ExitCode> {
    // Initialize logging
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive(tracing::Level::INFO.into()),
        )
        .init();

    // Parse CLI arguments
    let parsed = match cli::Cli::try_parse() {
        Ok(cli) => cli,
        Err(err) => {
            err.print().ok();
            return Ok(ExitCode::from(err.exit_code().try_into().unwrap_or(2)));
        }
    };

    let runtime = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()?;
    let result = runtime.block_on(parsed.execute());

    match result {
        Ok(exit_code) => Ok(exit_code),
        Err(err) => {
            eprintln!("{} {err}", "error:".bold().red());
            Ok(ExitCode::FAILURE)
        }
    }
}
