use std::path::PathBuf;

use tracing::debug;

use crate::cli::chat::tools;
use crate::cli::chat::tools::Tool;

use super::super::super::ChatState;
use super::doc_gen::generate_or_update_readme;

pub async fn execute(path: &str) -> ChatState {
    let scan_tool = tools::IgnoreScanTool;
    let args = serde_json::json!({ "path": path });
    let dir = match scan_tool.invoke(args).await {
        Ok(d) => d,
        Err(e) => {
            println!("Error during directory scan: {}", e);
            return ChatState::PromptUser {
                skip_printing_tools: false,
            };
        }
    };

    debug!("Directory scan result: {:?}", dir);

    let readme_path = PathBuf::from(path).join("README.md");
    let existing_readme = std::fs::read_to_string(&readme_path).ok();

    if existing_readme.is_some() {
        println!("Found existing README.md — improving it...");
    } else {
        println!("No README.md found — generating from scratch...");
    }

    let response = generate_or_update_readme(&dir, existing_readme.as_deref()).await;

    let readme_content = extract_readme_content(&response);

    let readme_path = PathBuf::from(path).join("README.md");
    match std::fs::write(&readme_path, &readme_content) {
        Ok(_) => println!("✓ README.md written to {}", readme_path.display()),
        Err(e) => println!("Failed to write README.md: {}", e),
    }

    ChatState::PromptUser {
        skip_printing_tools: true,
    }
}

fn extract_readme_content(response: &str) -> String {
    if let (Some(start), Some(end)) = (response.find("<readme>"), response.find("</readme>")) {
        response[start + "<readme>".len()..end].trim().to_string()
    } else {
        response.trim().to_string()
    }
}
