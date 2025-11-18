pub mod api;
pub mod cli;
pub mod grpc;
pub mod util;

/// Cleanup function to be called when the program exits
/// This ensures all resources (like gRPC server) are properly shut down
pub async fn cleanup() {
    cli::chat::cli::slash_commands::shutdown_grpc_server().await;
}
