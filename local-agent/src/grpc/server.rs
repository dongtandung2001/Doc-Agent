use tonic::transport::Server;
use tokio::task::JoinHandle;
use std::net::SocketAddr;

use crate::grpc::proto::local_agent_service_server::LocalAgentServiceServer;
use crate::grpc::service::LocalAgentServiceImpl;

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