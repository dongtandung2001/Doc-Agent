use crate::cli::chat::message::Message;
use crate::cli::chat::tool_manager::ToolManager;
use crate::cli::chat::ToolUse;

use serde_json::json;

#[derive(Debug, Clone)]
pub struct FileReadResult {
    pub path: String,
    pub content: Option<String>,
    pub error: Option<String>,
}

pub async fn read_file_content(paths: Vec<String>) -> Vec<FileReadResult> {
    let tool_manager = ToolManager::new(true);
    let mut tools_to_execute = Vec::new();

    for p in &paths {
        let tool_use = ToolUse {
            id: "fs_read".to_string(),
            name: "fs_read".to_string(),
            args: json!({
                "path": p
            }),
        };
        tools_to_execute.push(tool_use);
    }

    let res = tool_manager.execute_all(tools_to_execute).await;

    // Convert Message results to FileReadResult
    paths
        .into_iter()
        .zip(res.into_iter())
        .map(|(path, message)| match message {
            Message::Tool { content, .. } => {
                if content.starts_with("Error:") {
                    FileReadResult {
                        path,
                        content: None,
                        error: Some(content),
                    }
                } else {
                    FileReadResult {
                        path,
                        content: Some(content),
                        error: None,
                    }
                }
            }
            _ => FileReadResult {
                path,
                content: None,
                error: Some("Unexpected message type".to_string()),
            },
        })
        .collect()
}
