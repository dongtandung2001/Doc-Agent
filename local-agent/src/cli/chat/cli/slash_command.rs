use super::super::ChatState;

/// Slash commands for chat
#[derive(Debug, Clone)]
pub enum SlashCommand {
    Quit,
}

impl SlashCommand {
    /// Parse input as slash command
    pub fn parse(input: &str) -> Option<Self> {
        let trimmed = input.trim();

        if !trimmed.starts_with('/') {
            return None;
        }

        match trimmed {
            "/quit" | "/exit" | "/q" => Some(Self::Quit),
            _ => None,
        }
    }

    /// Execute command and return next state
    pub fn execute(&self) -> ChatState {
        match self {
            Self::Quit => {
                println!("Exiting...");
                ChatState::Exit
            }
        }
    }
}
