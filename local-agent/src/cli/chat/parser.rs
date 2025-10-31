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

/// Parse incremental response (DEPRECATED - kept for future use)
#[allow(dead_code)]
pub fn parse_incremental_response(response: Response) -> impl Stream<Item = Result<ResponseEvent>> {
    response
        .bytes_stream()
        .map_err(|e| eyre::eyre!("Response error: {}", e))
        .map(|result| {
            result.and_then(|chunk| {
                let text = String::from_utf8_lossy(&chunk);
                // eprintln!("[DEBUG] Received chunk: {}", text);
                parse_response_chunk(&text)
            })
        })
        .flat_map(|result| {
            futures::stream::iter(match result {
                Ok(events) => events.into_iter().map(Ok).collect::<Vec<_>>(),
                Err(e) => vec![Err(e)],
            })
        })
}

fn parse_response_chunk(text: &str) -> Result<Vec<ResponseEvent>> {
    let mut events = Vec::new();

    for line in text.lines() {
        if let Some(data) = line.strip_prefix("data: ") {
            eprintln!("[DEBUG] Processing data line: '{}'", data);

            // Handle [DONE] marker
            if data.trim() == "[DONE]" {
                eprintln!("[DEBUG] Found [DONE] marker");
                events.push(ResponseEvent::Done);
                continue;
            }

            // Skip empty data lines
            let data = data.trim();
            if data.is_empty() {
                eprintln!("[DEBUG] Skipping empty data line");
                continue;
            }

            // Parse JSON data
            let event: Value = serde_json::from_str(data)?;

            eprintln!("[DEBUG] Parsed event: {}", serde_json::to_string(&event)?);

            // Handle OpenAI-style streaming format
            if let Some(choices) = event.get("choices").and_then(|c| c.as_array()) {
                if let Some(choice) = choices.first() {
                    // Check for finish_reason to detect end of stream
                    if let Some(finish_reason) = choice.get("finish_reason") {
                        if !finish_reason.is_null() {
                            eprintln!("[DEBUG] Response finished with reason: {:?}", finish_reason);
                            events.push(ResponseEvent::Done);
                            continue;
                        }
                    }

                    // Handle delta content
                    if let Some(delta) = choice.get("delta") {
                        // Text content delta
                        if let Some(content) = delta.get("content").and_then(|c| c.as_str()) {
                            events.push(ResponseEvent::ContentDelta {
                                text: content.to_string(),
                            });
                        }

                        // Tool calls delta (OpenAI format)
                        if let Some(tool_calls) = delta.get("tool_calls").and_then(|t| t.as_array())
                        {
                            eprintln!("[DEBUG] Found tool_calls in delta: {:?}", tool_calls);

                            if let Some(tool_call) = tool_calls.first() {
                                eprintln!("[DEBUG] First tool_call: {:?}", tool_call);

                                // Check for tool call ID (indicates start of new tool)
                                if let Some(id) = tool_call.get("id").and_then(|i| i.as_str()) {
                                    eprintln!("[DEBUG] Found tool ID: {}", id);

                                    // Get function info
                                    if let Some(function) = tool_call.get("function") {
                                        eprintln!("[DEBUG] Function object: {:?}", function);

                                        if let Some(name) =
                                            function.get("name").and_then(|n| n.as_str())
                                        {
                                            eprintln!("[DEBUG] Tool name: {}", name);
                                            events.push(ResponseEvent::ToolUseStart {
                                                id: id.to_string(),
                                                name: name.to_string(),
                                            });
                                        }
                                    }
                                }

                                // Tool arguments delta (might be in same or different chunk)
                                if let Some(function) = tool_call.get("function") {
                                    if let Some(arguments) =
                                        function.get("arguments").and_then(|a| a.as_str())
                                    {
                                        eprintln!("[DEBUG] Found arguments delta: {}", arguments);
                                        events.push(ResponseEvent::ToolInputDelta {
                                            json: arguments.to_string(),
                                        });
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    Ok(events)
}
