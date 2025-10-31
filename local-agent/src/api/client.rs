use eyre::Result;
use reqwest::{Client, Response};
use serde_json::{json, Value};

/// API client for chat
pub struct ApiClient {
    client: Client,
    endpoint: String,
    api_key: String,
    model: String,
}

impl ApiClient {
    pub fn new(endpoint: &str, api_key: &str, model: &str) -> Result<Self> {
        Ok(Self {
            client: Client::new(),
            endpoint: endpoint.to_string(),
            api_key: api_key.to_string(),
            model: model.to_string(),
        })
    }

    /// Send chat request and return JSON response
    pub async fn send_message(&self, messages: Vec<Value>, tools: Vec<Value>) -> Result<Value> {
        // eprintln!("[DEBUG] ApiClient - Endpoint: {}", self.endpoint);
        // eprintln!("[DEBUG] ApiClient - Model: {}", self.model);
        // eprintln!("[DEBUG] ApiClient - API Key length: {}", self.api_key.len());

        // Build request body with required fields (no streaming)
        let request_body = json!({
            "model": self.model,
            "messages": messages,
            "stream": false,
            "tools": tools,
            "tool_choice": if tools.is_empty() { json!("none") } else { json!("auto") }
        });

        // eprintln!("[DEBUG] ApiClient - Request body: {}", serde_json::to_string_pretty(&request_body)?);

        let response = self
            .client
            .post(&self.endpoint)
            .header("Content-Type", "application/json")
            .header("Accept", "application/json")
            .header("Authorization", format!("Bearer {}", self.api_key))
            .json(&request_body)
            .send()
            .await?;

        if !response.status().is_success() {
            let status = response.status();
            let text = response.text().await?;
            eyre::bail!("API error {}: {}", status, text);
        }

        // Parse JSON response
        let json: Value = response.json().await?;
        eprintln!("[DEBUG] API Response: {}", serde_json::to_string_pretty(&json)?);

        Ok(json)
    }
}
