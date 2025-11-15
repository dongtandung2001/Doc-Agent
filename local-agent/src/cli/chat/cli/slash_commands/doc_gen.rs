use crate::cli::chat::tools::Tool;

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
    ChatState::PromptUser {
        skip_printing_tools: true,
    }
}


fn generate_readme() {
    
}