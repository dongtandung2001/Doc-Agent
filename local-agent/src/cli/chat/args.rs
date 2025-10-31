use clap::Args;
use eyre::Result;
use std::process::ExitCode;

use super::ChatSession;

/// Arguments for the chat subcommand
#[derive(Debug, Args)]
pub struct ChatArgs {
    /// Initial message to send
    pub input: Option<String>,

    /// Trust all tools (skip confirmation)
    #[arg(long)]
    pub trust_all_tools: bool,

    /// API endpoint URL
    #[arg(long, env = "CHAT_API_URL")]
    pub api_url: String,

    /// API key for authentication
    #[arg(long, env = "CHAT_API_KEY")]
    pub api_key: String,

    /// Model name to use
    #[arg(long, env = "CHAT_MODEL")]
    pub model: String,
}

impl Default for ChatArgs {
    fn default() -> Self {
        Self {
            input: None,
            trust_all_tools: false,
            api_url: std::env::var("CHAT_API_URL")
                .unwrap_or_else(|_| "https://api.token-ai.cn/v1/chat/completions".to_string()),
            api_key: std::env::var("CHAT_API_KEY")
                .unwrap_or_else(|_| String::new()),
            model: std::env::var("CHAT_MODEL")
                .unwrap_or_else(|_| "DeepSeek-V3".to_string()),
        }
    }
}

impl ChatArgs {
    pub async fn execute(self) -> Result<ExitCode> {
        let mut session = ChatSession::new(self).await?;
        session.spawn().await?;
        Ok(ExitCode::SUCCESS)
    }
}
