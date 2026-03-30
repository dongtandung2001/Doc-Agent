use clap::Parser;
use crossterm::style::Stylize;
use eyre::Result;
use std::process::ExitCode;

use local_agent::cli;
use local_agent::grpc::server;

fn main() -> Result<ExitCode> {
    // Load .env file if it exists (ignore if not found)
    dotenvy::dotenv().ok();

    // Initialize logging — all output goes to a rolling log file, nothing to terminal
    let file_appender = tracing_appender::rolling::daily("logs", "local-agent.log");
    let (non_blocking, _guard) = tracing_appender::non_blocking(file_appender);
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive(tracing::Level::DEBUG.into()),
        )
        .with_writer(non_blocking)
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

    let result = runtime.block_on(async {
        let exit_result = parsed.execute().await;

        // Always cleanup gRPC server before exiting
        if let Err(e) = server::shutdown_global_server().await {
            tracing::error!("Error shutting down gRPC server: {}", e);
        }

        exit_result
    });

    match result {
        Ok(exit_code) => Ok(exit_code),
        Err(err) => {
            eprintln!("{} {err}", "error:".bold().red()); // user-facing fatal error
            Ok(ExitCode::FAILURE)
        }
    }
}
