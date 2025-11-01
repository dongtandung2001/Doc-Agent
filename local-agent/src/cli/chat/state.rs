use serde_json::Value;

/// Represents the current state of the chat session
#[derive(Debug)]
pub enum ChatState {
    /// Waiting for user input
    PromptUser {
        skip_printing_tools: bool,
    },

    /// Processing user input
    HandleInput {
        input: String,
    },

    /// Handling AI response
    HandleResponse {
        request: String,
    },

    /// Validating tool permissions
    ValidateTools {
        tools: Vec<ToolUse>,
    },

    /// Executing tools
    ExecuteTools {
        tools: Vec<ToolUse>,
    },

    /// Exit the chat
    Exit,
}

impl Default for ChatState {
    fn default() -> Self {
        Self::PromptUser {
            skip_printing_tools: false,
        }
    }
}

#[derive(Debug, Clone)]
pub struct ToolUse {
    pub id: String,
    pub name: String,
    pub args: Value,
}
