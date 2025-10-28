use eyre::Result;
use futures::stream::{Stream, StreamExt, TryStreamExt};
use reqwest::Response;
use serde_json::Value;

/// Events from the streaming API
#[derive(Debug, Clone)]
pub enum StreamEvent {
    /// Text content delta
    ContentDelta { text: String },

    /// Tool use block started
    ToolUseStart {
        id: String,
        name: String,
    },

    /// Tool use input delta
    ToolInputDelta { json: String },

    /// Stream complete
    Done,
}

/// Parse Server-Sent Events stream
pub fn parse_sse_stream(response: Response) -> impl Stream<Item = Result<StreamEvent>> {
    response
        .bytes_stream()
        .map_err(|e| eyre::eyre!("Stream error: {}", e))
        .and_then(|chunk| async move {
            let text = String::from_utf8_lossy(&chunk);
            parse_sse_chunk(&text)
        })
        .filter_map(|result| async move {
            match result {
                Ok(Some(event)) => Some(Ok(event)),
                Ok(None) => None,
                Err(e) => Some(Err(e)),
            }
        })
}

fn parse_sse_chunk(text: &str) -> Result<Option<StreamEvent>> {
    for line in text.lines() {
        if let Some(data) = line.strip_prefix("data: ") {
            if data == "[DONE]" {
                return Ok(Some(StreamEvent::Done));
            }

            let event: Value = serde_json::from_str(data)?;

            // Parse different event types based on the API response format
            if let Some(event_type) = event.get("type").and_then(|t| t.as_str()) {
                match event_type {
                    "content_block_delta" => {
                        if let Some(delta) = event.get("delta") {
                            // Check for text delta
                            if let Some(text) = delta.get("text").and_then(|t| t.as_str()) {
                                return Ok(Some(StreamEvent::ContentDelta {
                                    text: text.to_string(),
                                }));
                            }
                            // Check for tool input delta
                            if delta.get("type").and_then(|t| t.as_str()) == Some("input_json_delta") {
                                if let Some(partial_json) = delta.get("partial_json").and_then(|j| j.as_str()) {
                                    return Ok(Some(StreamEvent::ToolInputDelta {
                                        json: partial_json.to_string(),
                                    }));
                                }
                            }
                        }
                    }
                    "content_block_start" => {
                        if let Some(content_block) = event.get("content_block") {
                            if content_block.get("type").and_then(|t| t.as_str()) == Some("tool_use") {
                                let id = content_block.get("id")
                                    .and_then(|i| i.as_str())
                                    .unwrap_or("")
                                    .to_string();
                                let name = content_block.get("name")
                                    .and_then(|n| n.as_str())
                                    .unwrap_or("")
                                    .to_string();
                                return Ok(Some(StreamEvent::ToolUseStart { id, name }));
                            }
                        }
                    }
                    "message_stop" => {
                        return Ok(Some(StreamEvent::Done));
                    }
                    _ => {}
                }
            }
        }
    }

    Ok(None)
}
