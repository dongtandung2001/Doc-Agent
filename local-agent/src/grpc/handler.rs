use crate::cli::chat::message::Message;
use crate::cli::chat::tool_manager::ToolManager;
use crate::cli::chat::ToolUse;

use serde_json::json;

use crate::grpc::proto::FileReadArg;

#[derive(Debug, Clone)]
pub struct FileReadResult {
    pub id: String,
    pub content: Option<String>,
}

pub async fn read_file_content(args: Vec<FileReadArg>) -> Vec<FileReadResult> {
    let tool_manager = ToolManager::new(true);
    let mut tools_to_execute = Vec::new();

    for p in &args {
        let tool_use = ToolUse {
            id: p.id.clone(),
            name: "fs_read".to_string(),
            args: json!({
                "path": p.path
            }),
        };
        tools_to_execute.push(tool_use);
    }

    let res = tool_manager.execute_all(tools_to_execute).await;

    // Convert Message results to FileReadResult
    args.into_iter()
        .zip(res.into_iter())
        .map(|(arg, message)| match message {
            Message::Tool { content, .. } => FileReadResult {
                id: arg.id,
                content: Some(content),
            },
            _ => FileReadResult {
                id: arg.id,
                content: Some("Unexpected message type".to_string()),
            },
        })
        .collect()
}
