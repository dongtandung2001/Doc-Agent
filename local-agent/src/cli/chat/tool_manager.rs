use eyre::Result;
use serde_json::Value;

use super::tools::ToolRegistry;
use super::state::ToolUse;
use super::message::Message;

/// Manages tool execution and results
pub struct ToolManager {
    registry: ToolRegistry,
}

impl ToolManager {
    pub fn new(trust_all: bool) -> Self {
        Self {
            registry: ToolRegistry::new(trust_all),
        }
    }

    /// Get tool specs for API
    pub fn tool_specs(&self) -> Vec<serde_json::Value> {
        self.registry
            .all_specs()
            .into_iter()
            .map(|spec| {
                serde_json::json!({
                    "name": spec.name,
                    "description": spec.description,
                    "input_schema": spec.input_schema,
                })
            })
            .collect()
    }

    /// Check if tools require approval
    pub fn requires_approval(&self, tools: &[ToolUse]) -> bool {
        tools.iter().any(|t| self.registry.requires_approval(&t.name))
    }

    /// Get preview of tool execution
    pub fn preview_tools(&self, tools: &[ToolUse]) -> Vec<String> {
        tools
            .iter()
            .map(|t| {
                if let Some(tool) = self.registry.get(&t.name) {
                    tool.preview(&t.args)
                } else {
                    format!("Unknown tool: {}", t.name)
                }
            })
            .collect()
    }

    /// Execute a single tool
    pub async fn execute_tool(&self, tool_use: &ToolUse) -> Message {
        match self.registry.get(&tool_use.name) {
            Some(tool) => {
                match tool.invoke(tool_use.args.clone()).await {
                    Ok(result) => Message::tool_result(&tool_use.id, result, false),
                    Err(err) => {
                        let error_msg = format!("Error: {}", err);
                        Message::tool_result(&tool_use.id, error_msg, true)
                    }
                }
            }
            None => {
                let error_msg = format!("Unknown tool: {}", tool_use.name);
                Message::tool_result(&tool_use.id, error_msg, true)
            }
        }
    }

    /// Execute all tools and return results
    pub async fn execute_all(&self, tools: Vec<ToolUse>) -> Vec<Message> {
        let mut results = Vec::new();
        for tool_use in tools {
            results.push(self.execute_tool(&tool_use).await);
        }
        results
    }
}
