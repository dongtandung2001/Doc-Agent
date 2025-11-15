mod doc_gen;
mod quit;

use super::super::ChatState;

/// Parse and execute a slash command
/// Returns Some(ChatState) if the input is a valid slash command, None otherwise
pub async fn parse_and_execute(input: &str, root_dir: &str) -> Option<ChatState> {
    let trimmed = input.trim();

    // Must start with '/'
    if !trimmed.starts_with('/') {
        return None;
    }

    // Remove the '/' prefix and match the command
    let command_name = &trimmed[1..];

    match command_name {
        "quit" | "exit" | "q" => Some(quit::execute()),
        "init" | "start" | "doc_generation" => Some(doc_gen::execute(root_dir).await),
        _ => None,
    }
}
