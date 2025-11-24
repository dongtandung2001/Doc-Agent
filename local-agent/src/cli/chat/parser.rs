use eyre::Result;
use futures::stream::{Stream, StreamExt, TryStreamExt};
use reqwest::Response;
use serde_json::Value;

use super::state::ToolUse;

/// Parsed response from the API
#[derive(Debug, Clone)]
pub struct ParsedResponse {
    /// Assistant's text response (if any)
    pub assistant_text: String,
    /// Tool calls to execute (if any)
    pub tools_to_execute: Vec<ToolUse>,
}

/// Parse a complete API response (non-incremental mode)
pub fn parse_response(response: &Value) -> Result<ParsedResponse> {
    let mut assistant_text = String::new();
    let mut tools_to_execute = Vec::new();

    // Extract from choices[0].message
    if let Some(choices) = response.get("choices").and_then(|c| c.as_array()) {
        if let Some(choice) = choices.first() {
            if let Some(message) = choice.get("message") {
                // Extract text content
                if let Some(content) = message.get("content").and_then(|c| c.as_str()) {
                    assistant_text = content.to_string();
                }

                // finish reason
                if let Some(finish_reason) = choice.get("finish_reason").and_then(|f| f.as_str()) {
                    eprintln!("[DEBUG] Finish reason: {}", finish_reason);
                }

                // Extract tool calls
                if let Some(tool_calls) = message.get("tool_calls").and_then(|t| t.as_array()) {
                    eprintln!("[DEBUG] Found {} tool calls", tool_calls.len());

                    for tool_call in tool_calls {
                        if let Some(function) = tool_call.get("function") {
                            let id = tool_call
                                .get("id")
                                .and_then(|i| i.as_str())
                                .unwrap_or("")
                                .to_string();

                            let name = function
                                .get("name")
                                .and_then(|n| n.as_str())
                                .unwrap_or("")
                                .to_string();

                            let arguments = function
                                .get("arguments")
                                .and_then(|a| a.as_str())
                                .unwrap_or("{}");

                            eprintln!("[DEBUG] Tool: {} with args: {}", name, arguments);

                            let args: Value = serde_json::from_str(arguments)?;

                            tools_to_execute.push(ToolUse { id, name, args });
                        }
                    }
                }
            }
        }
    }

    Ok(ParsedResponse {
        assistant_text,
        tools_to_execute,
    })
}

/// Events from the API response (DEPRECATED - kept for future use)
#[derive(Debug, Clone)]
#[allow(dead_code)]
pub enum ResponseEvent {
    /// Text content delta
    ContentDelta { text: String },

    /// Tool use block started
    ToolUseStart { id: String, name: String },

    /// Tool use input delta
    ToolInputDelta { json: String },

    /// Response complete
    Done,
}
