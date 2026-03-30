use super::handler;
use crate::grpc::proto::{
    local_agent_service_server::LocalAgentService, FileReadResult, HealthCheckRequest,
    HealthCheckResponse, RequestFileContentRequest, RequestFileContentResponse,
};
use tonic::{Request, Response, Status};
use tracing::debug;

#[derive(Debug, Default)]
pub struct LocalAgentServiceImpl;

#[tonic::async_trait]
impl LocalAgentService for LocalAgentServiceImpl {
    async fn request_file_content(
        &self,
        request: Request<RequestFileContentRequest>,
    ) -> Result<Response<RequestFileContentResponse>, Status> {
        let req = request.into_inner();
        let arg = req.args;
        let file_read_results = handler::read_file_content(arg).await;

        debug!("Read file content results: {:?}", file_read_results);

        // Convert handler::FileReadResult to proto::FileReadResult
        let final_results: Vec<FileReadResult> = file_read_results
            .into_iter()
            .map(|res| FileReadResult {
                id: res.id,
                content: res.content.unwrap_or_default(),
            })
            .collect();
        debug!("Final results: {} files", final_results.len());
        Ok(Response::new(RequestFileContentResponse {
            results: final_results,
        }))
    }

    async fn health_check(
        &self,
        _request: Request<HealthCheckRequest>,
    ) -> Result<Response<HealthCheckResponse>, Status> {
        debug!("Health check received");
        Ok(Response::new(HealthCheckResponse { is_alive: true }))
    }
}
