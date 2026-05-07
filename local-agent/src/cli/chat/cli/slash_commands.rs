mod doc_gen;
mod quit;
mod readme;

use crate::api::ApiClient;
use tracing::debug;

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

        // Remove the '/' prefix
        let body = &trimmed[1..];

        // Split into the "verb" (e.g., 'init') and the "args" (e.g., './path')
        // split_once returns Some((left, right)) if a space is found, else None
        let (cmd, args) = match body.split_once(' ') {
            Some((v, a)) => (v, Some(a.trim())), // Found a space, trim the args
            None => (body, None),                // No space, the whole body is the verb
        };

        debug!("Command: {}, Args: {:?}, Root: {}", cmd, args, root_dir);

        match cmd {
            "quit" | "exit" | "q" => Some(quit::execute()),

            "init" | "start" | "doc_generation" => {
                // If the user provided args, use them; otherwise use root_dir
                let path = args.unwrap_or(root_dir);
                debug!("Using path: {}", path);
                Some(doc_gen::execute(path, api_client).await)
            }

            "readme" => {
                let path = args.unwrap_or(root_dir);
                debug!("Using path: {}", path);
                Some(readme::execute(path).await)
            }

            _ => None,
        }
    })
}
