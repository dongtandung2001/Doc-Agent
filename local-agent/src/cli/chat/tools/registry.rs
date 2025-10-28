use async_trait::async_trait;
use eyre::Result;
use serde_json::Value;
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct ToolSpec {
    pub name: String,
    pub description: String,
    pub input_schema: Value,
}

#[async_trait]
pub trait Tool: Send + Sync {
    fn name(&self) -> &str;
    fn description(&self) -> &str;
    fn input_schema(&self) -> Value;
    async fn invoke(&self, args: Value) -> Result<String>;

    fn preview(&self, args: &Value) -> String {
        format!("{} with {}", self.name(), args)
    }

    fn spec(&self) -> ToolSpec {
        ToolSpec {
            name: self.name().to_string(),
            description: self.description().to_string(),
            input_schema: self.input_schema(),
        }
    }
}

pub struct ToolRegistry {
    tools: HashMap<String, Box<dyn Tool>>,
    trust_all: bool,
}

impl ToolRegistry {
    pub fn new(trust_all: bool) -> Self {
        let mut registry = Self {
            tools: HashMap::new(),
            trust_all,
        };

        registry.register(Box::new(super::fs_read::FsReadTool));
        registry.register(Box::new(super::fs_scan::FsScanTool));

        registry
    }

    pub fn register(&mut self, tool: Box<dyn Tool>) {
        self.tools.insert(tool.name().to_string(), tool);
    }

    pub fn get(&self, name: &str) -> Option<&Box<dyn Tool>> {
        self.tools.get(name)
    }

    pub fn all_specs(&self) -> Vec<ToolSpec> {
        self.tools.values().map(|t| t.spec()).collect()
    }

    pub fn requires_approval(&self, name: &str) -> bool {
        if self.trust_all {
            false
        } else {
            // fs_read is automatically trusted, others require approval
            name != "fs_read"
        }
    }
}
