use clap::Args;
use eyre::Result;
use std::process::ExitCode;

use super::ChatSession;

/// Arguments for the chat subcommand
#[derive(Debug, Args, Default)]
pub struct ChatArgs {
    /// Initial message to send
    pub input: Option<String>,

    /// Trust all tools (skip confirmation)
    #[arg(long)]
    pub trust_all_tools: bool,

    /// API endpoint URL
    #[arg(long, env = "CHAT_API_URL", default_value = "http://localhost:8080")]
    pub api_url: String,
}

impl ChatArgs {
    pub async fn execute(self) -> Result<ExitCode> {
        let mut session = ChatSession::new(self).await?;
        session.spawn().await?;
        Ok(ExitCode::SUCCESS)
    }
}
