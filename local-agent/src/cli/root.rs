use clap::{Parser, Subcommand};
use eyre::Result;
use std::process::ExitCode;

use super::chat::ChatArgs;

/// Simple AI chat CLI
#[derive(Debug, Parser)]
#[command(name = "local-agent")]
#[command(about = "AI-powered chat assistant with tools")]
pub struct Cli {
    #[command(subcommand)]
    pub subcommand: Option<RootSubcommand>,
}

/// Root-level subcommands
#[derive(Debug, Subcommand)]
pub enum RootSubcommand {
    /// Start an interactive chat session
    Chat(ChatArgs),
}

impl Default for RootSubcommand {
    fn default() -> Self {
        Self::Chat(ChatArgs::default())
    }
}

impl Cli {
    pub async fn execute(self) -> Result<ExitCode> {
        let subcommand = self.subcommand.unwrap_or_default();
        subcommand.execute().await
    }
}

impl RootSubcommand {
    pub async fn execute(self) -> Result<ExitCode> {
        match self {
            Self::Chat(args) => args.execute().await,
        }
    }
}
