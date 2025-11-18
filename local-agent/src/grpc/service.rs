use super::handler;
use crate::grpc::proto::{
    local_agent_service_server::LocalAgentService, HealthCheckRequest, HealthCheckResponse,
    RequestFileContentRequest, RequestFileContentResponse,
};
use tonic::{Request, Response, Status};

#[derive(Debug, Default)]
pub struct LocalAgentServiceImpl;

#[tonic::async_trait]
impl LocalAgentService for LocalAgentServiceImpl {
    async fn request_file_content(
        &self,
        request: Request<RequestFileContentRequest>,
    ) -> Result<Response<RequestFileContentResponse>, Status> {
        let req = request.into_inner();
        let paths = req.paths;
        let results = handler::read_file_content(paths).await;

        println!("Read file content results: {:?}", results);
        Ok(Response::new(RequestFileContentResponse {
            content: results
                .into_iter()
                .map(|res| match res.error {
                    Some(err) => format!("Error reading {}: {}", res.path, err),
                    None => res.content.unwrap_or_default(),
                })
                .collect(),
        }))
    }

    async fn health_check(
        &self,
        _request: Request<HealthCheckRequest>,
    ) -> Result<Response<HealthCheckResponse>, Status> {
        println!("Health check received");
        Ok(Response::new(HealthCheckResponse { is_alive: true }))
    }
}
