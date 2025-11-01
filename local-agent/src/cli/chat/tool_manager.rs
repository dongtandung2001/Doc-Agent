use eyre::Result;
use serde_json::Value;
use std::sync::Arc;
use std::time::Instant;
use tracing::{debug, info};

use super::message::Message;
use super::state::ToolUse;
use super::tools::ToolRegistry;

/// Manages tool execution and results
pub struct ToolManager {
    registry: Arc<ToolRegistry>,
}

impl ToolManager {
    pub fn new(trust_all: bool) -> Self {
        Self {
            registry: Arc::new(ToolRegistry::new(trust_all)),
        }
    }

    /// Get tool specs for API (OpenAI function calling format)
    pub fn tool_specs(&self) -> Vec<serde_json::Value> {
        self.registry
            .all_specs()
            .into_iter()
            .map(|spec| {
                serde_json::json!({
                    "type": "function",
                    "function": {
                        "name": spec.name,
                        "description": spec.description,
                        "parameters": spec.input_schema,
                    }
                })
            })
            .collect()
    }

    /// Check if tools require approval
    pub fn requires_approval(&self, tools: &[ToolUse]) -> bool {
        tools
            .iter()
            .any(|t| self.registry.requires_approval(&t.name))
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
            Some(tool) => match tool.invoke(tool_use.args.clone()).await {
                Ok(result) => Message::tool_result(&tool_use.id, result),
                Err(err) => {
                    let error_msg = format!("Error: {}", err);
                    Message::tool_result(&tool_use.id, error_msg)
                }
            },
            None => {
                let error_msg = format!("Unknown tool: {}", tool_use.name);
                Message::tool_result(&tool_use.id, error_msg)
            }
        }
    }

    /// Execute all tools in parallel and return results
    pub async fn execute_all(&self, tools: Vec<ToolUse>) -> Vec<Message> {
        let total_start = Instant::now();
        let tool_count = tools.len();

        debug!("Starting parallel execution of {} tools", tool_count);

        // Spawn parallel tasks for each tool
        let handles: Vec<_> = tools
            .into_iter()
            .map(|tool_use| {
                let registry = Arc::clone(&self.registry);
                tokio::spawn(async move { Self::execute_tool_internal(registry, tool_use).await })
            })
            .collect();

        // Collect results, handling any panics in spawned tasks
        let mut results = Vec::with_capacity(handles.len());
        for handle in handles {
            match handle.await {
                Ok(message) => results.push(message),
                Err(err) => {
                    // Task panicked - create error message
                    let error_msg = format!("Task execution failed: {}", err);
                    results.push(Message::tool_result("error", error_msg));
                }
            }
        }

        let total_elapsed = total_start.elapsed();
        info!(
            "Completed {} tools in {:.2}ms (parallel execution)",
            tool_count,
            total_elapsed.as_secs_f64() * 1000.0
        );

        results
    }

    /// Internal helper to execute a tool (used by parallel execution)
    async fn execute_tool_internal(registry: Arc<ToolRegistry>, tool_use: ToolUse) -> Message {
        let start = Instant::now();
        let tool_name = tool_use.name.clone();
        let tool_id = tool_use.id.clone();

        debug!(
            "Starting execution of tool: {} (id: {})",
            tool_name, tool_id
        );

        let message = match registry.get(&tool_use.name) {
            Some(tool) => match tool.invoke(tool_use.args.clone()).await {
                Ok(result) => Message::tool_result(&tool_use.id, result),
                Err(err) => {
                    let error_msg = format!("Error: {}", err);
                    Message::tool_result(&tool_use.id, error_msg)
                }
            },
            None => {
                let error_msg = format!("Unknown tool: {}", tool_use.name);
                Message::tool_result(&tool_use.id, error_msg)
            }
        };

        let elapsed = start.elapsed();
        info!(
            "Tool '{}' (id: {}) completed in {:.2}ms",
            tool_name,
            tool_id,
            elapsed.as_secs_f64() * 1000.0
        );

        message
    }
}
