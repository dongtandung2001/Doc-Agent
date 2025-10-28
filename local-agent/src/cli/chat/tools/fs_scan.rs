use async_trait::async_trait;
use eyre::Result;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use tokio::fs;

use super::registry::Tool;

pub struct FsScanTool;

#[derive(Debug, Deserialize, Serialize)]
struct FsScanArgs {
    path: String,
    #[serde(default = "default_depth")]
    max_depth: usize,
}

fn default_depth() -> usize { 1 }

#[async_trait]
impl Tool for FsScanTool {
    fn name(&self) -> &str {
        "fs_scan"
    }

    fn description(&self) -> &str {
        "List directory contents with depth control"
    }

    fn input_schema(&self) -> Value {
        json!({
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": "Directory to scan"
                },
                "max_depth": {
                    "type": "integer",
                    "description": "Max depth (default: 1)"
                }
            },
            "required": ["path"]
        })
    }

    async fn invoke(&self, args: Value) -> Result<String> {
        let args: FsScanArgs = serde_json::from_value(args)?;
        let mut result = Vec::new();
        scan_dir(&args.path, 0, args.max_depth, &mut result).await?;
        Ok(result.join("\n"))
    }

    fn preview(&self, args: &Value) -> String {
        if let Ok(args) = serde_json::from_value::<FsScanArgs>(args.clone()) {
            format!("Scan {} (depth: {})", args.path, args.max_depth)
        } else {
            "Scan directory (invalid args)".to_string()
        }
    }
}

fn scan_dir<'a>(
    path: &'a str,
    depth: usize,
    max_depth: usize,
    result: &'a mut Vec<String>
) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<()>> + Send + 'a>> {
    Box::pin(async move {
        if depth > max_depth {
            return Ok(());
        }

        let mut entries = fs::read_dir(path).await?;
        let indent = "  ".repeat(depth);

        while let Some(entry) = entries.next_entry().await? {
            let path = entry.path();
            let name = entry.file_name();
            let metadata = entry.metadata().await?;

            let icon = if metadata.is_dir() { "[DIR]" } else { "[FILE]" };
            result.push(format!("{}{} {}", indent, icon, name.to_string_lossy()));

            if metadata.is_dir() && depth < max_depth {
                scan_dir(path.to_str().unwrap(), depth + 1, max_depth, result).await?;
            }
        }

        Ok(())
    })
}
