use serde::{Deserialize, Serialize};
use serde_json::Value;

/// Represents a message in the conversation
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "role")]
pub enum Message {
    #[serde(rename = "user")]
    User { content: String },

    #[serde(rename = "assistant")]
    Assistant { content: String },

    #[serde(rename = "tool_use")]
    ToolUse {
        id: String,
        name: String,
        args: Value,
    },

    #[serde(rename = "tool_result")]
    ToolResult {
        tool_use_id: String,
        content: String,
        is_error: bool,
    },
}

impl Message {
    pub fn user(content: impl Into<String>) -> Self {
        Self::User {
            content: content.into(),
        }
    }

    pub fn assistant(content: impl Into<String>) -> Self {
        Self::Assistant {
            content: content.into(),
        }
    }

    pub fn tool_use(id: impl Into<String>, name: impl Into<String>, args: Value) -> Self {
        Self::ToolUse {
            id: id.into(),
            name: name.into(),
            args,
        }
    }

    pub fn tool_result(tool_use_id: impl Into<String>, content: impl Into<String>, is_error: bool) -> Self {
        Self::ToolResult {
            tool_use_id: tool_use_id.into(),
            content: content.into(),
            is_error,
        }
    }
}
