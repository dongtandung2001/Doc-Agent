use tonic::transport::Server;
use tokio::task::JoinHandle;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::Mutex;
use once_cell::sync::Lazy;

use crate::grpc::proto::local_agent_service_server::LocalAgentServiceServer;
use crate::grpc::service::LocalAgentServiceImpl;

// Global server handle to keep the gRPC server alive
static GLOBAL_SERVER_HANDLE: Lazy<Arc<Mutex<Option<ServerHandle>>>> =
    Lazy::new(|| Arc::new(Mutex::new(None)));

pub struct ServerHandle {
    address: SocketAddr,
    shutdown_tx: tokio::sync::oneshot::Sender<()>,
    task_handle: JoinHandle<Result<(), tonic::transport::Error>>,
}

impl ServerHandle {
    pub fn address(&self) -> SocketAddr {
        self.address
    }

    pub async fn shutdown(self) -> Result<(), Box<dyn std::error::Error>> {
        // Send shutdown signal
        let _ = self.shutdown_tx.send(());

        // Wait for server to finish
        self.task_handle.await??;

        Ok(())
    }
}

/// Spawn gRPC server on the specified host and port
pub async fn spawn_server(
    host: &str,
    port: u16,
) -> Result<ServerHandle, Box<dyn std::error::Error>> {
    let addr = format!("{}:{}", host, port).parse::<SocketAddr>()?;

    let service = LocalAgentServiceImpl::default();
    let (shutdown_tx, shutdown_rx) = tokio::sync::oneshot::channel();

    let task_handle = tokio::spawn(async move {
        Server::builder()
            .add_service(LocalAgentServiceServer::new(service))
            .serve_with_shutdown(addr, async {
                shutdown_rx.await.ok();
            })
            .await
    });

    // Wait a bit to ensure server is up
    tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;

    Ok(ServerHandle {
        address: addr,
        shutdown_tx,
        task_handle,
    })
}

/// Check if the global gRPC server is already running
pub async fn is_server_running() -> bool {
    let handle_guard = GLOBAL_SERVER_HANDLE.lock().await;
    handle_guard.is_some()
}

/// Set the global gRPC server handle
pub async fn set_global_server(handle: ServerHandle) {
    *GLOBAL_SERVER_HANDLE.lock().await = Some(handle);
}

/// Shutdown the global gRPC server if it's running
pub async fn shutdown_global_server() -> Result<(), Box<dyn std::error::Error>> {
    let mut handle_guard = GLOBAL_SERVER_HANDLE.lock().await;

    if let Some(handle) = handle_guard.take() {
        println!("Shutting down gRPC server...");
        handle.shutdown().await?;
        println!("✓ gRPC server stopped successfully.");
        Ok(())
    } else {
        // Server not running, no output needed
        Ok(())
    }
}