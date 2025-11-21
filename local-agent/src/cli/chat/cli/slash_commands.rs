mod doc_gen;
mod quit;

use crate::api::ApiClient;

use super::super::ChatState;
use std::future::Future;
use std::pin::Pin;

/// Parse and execute a slash command
/// Returns Some(ChatState) if the input is a valid slash command, None otherwise
pub fn parse_and_execute<'a>(
    input: &'a str,
    root_dir: &'a str,
    api_client: &'a ApiClient,
) -> Pin<Box<dyn Future<Output = Option<ChatState>> + Send + 'a>> {
    Box::pin(async move {
        let trimmed = input.trim();

        // Must start with '/'
        if !trimmed.starts_with('/') {
            return None;
        }

        // Remove the '/' prefix and match the command
        let command_name = &trimmed[1..];

        match command_name {
            "quit" | "exit" | "q" => Some(quit::execute()),
            "init" | "start" | "doc_generation" => Some(doc_gen::execute(root_dir, api_client).await),
            _ => None,
        }
    })
}
