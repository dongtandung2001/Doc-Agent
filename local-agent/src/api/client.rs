use eyre::Result;
use httpmock::prelude::*;
use reqwest::{Client, Response};
use serde_json::{json, Value};

/// API client for chat
pub struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    pub fn new(base_url: &str) -> Result<Self> {
        Ok(Self {
            client: Client::new(),
            base_url: base_url.to_string(),
        })
    }

    /// Send chat request and return raw response
    pub async fn send_message(&self, messages: Vec<Value>, tools: Vec<Value>) -> Result<Response> {
        let request_body = json!({
            "messages": messages,
            "tools": tools,
            "stream": true,
        });

        let response = self
            .client
            .post(format!("{}/chat", self.base_url))
            .json(&request_body)
            .send()
            .await?;

        if !response.status().is_success() {
            let status = response.status();
            let text = response.text().await?;
            eyre::bail!("API error {}: {}", status, text);
        }

        Ok(response)
    }
}
