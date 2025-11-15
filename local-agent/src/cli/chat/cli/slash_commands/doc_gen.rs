use crate::cli::chat::tools::Tool;
use crate::cli::chat::ChatArgs;
use crate::cli::chat::ChatSession;

use super::super::super::ChatState;

use super::super::super::tools;

pub async fn execute(path: &str) -> ChatState {
    println!("Starting document generation...");
    println!("Please keep the agent running while documents are being generated.");

    // scan directory with gitignore support as part of doc generation
    let scanTool = tools::IgnoreScanTool;
    let args = serde_json::json!({ "path": path });
    let dir = scanTool.invoke(args).await;

    println!(
        "Directory scan completed. Generating documents...\n{:?}",
        dir
    );

    let chat_args = ChatArgs::default();

    let mut api_session = ChatSession::new(chat_args).await.unwrap();

    let response = api_session
        .send_message(String::from("Can you read this file: E:\\Code\\Doc-Agent\\local-agent\\src\\cli\\chat\\tools\\fs_read.rs"))
        .await
        .unwrap();

    println!("Received response from API: {}", response);
    ChatState::PromptUser {
        skip_printing_tools: true,
    }
}
