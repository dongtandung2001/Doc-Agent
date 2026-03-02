use async_trait::async_trait;
use eyre::{bail, Result};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use tokio::fs;

use super::registry::Tool;

pub struct FsReadTool;

#[derive(Debug, Deserialize, Serialize)]
struct FsReadArgs {
    path: String,
    #[serde(default)]
    start_line: Option<usize>,
    #[serde(default)]
    end_line: Option<usize>,
}

#[async_trait]
impl Tool for FsReadTool {
    fn name(&self) -> &str {
        "fs_read"
    }

    fn description(&self) -> &str {
        "Read file contents with optional line range"
    }

    fn input_schema(&self) -> Value {
        json!({
            "type": "object",
            "properties": {
                "path": {
                    "type": "string",
                    "description": "File path to read"
                },
                "start_line": {
                    "type": "integer",
                    "description": "Starting line (1-indexed)"
                },
                "end_line": {
                    "type": "integer",
                    "description": "Ending line (1-indexed, inclusive)"
                }
            },
            "required": ["path"]
        })
    }

    async fn invoke(&self, args: Value) -> Result<String> {
        let args: FsReadArgs = serde_json::from_value(args)?;

        let metadata = fs::metadata(&args.path)
            .await
            .map_err(|e| eyre::eyre!("Cannot access '{}': {}", args.path, e))?;
        if metadata.is_dir() {
            bail!(
                "'{}' is a directory, not a file. Use fs_scan to list its contents.",
                args.path
            );
        }

        let content = fs::read_to_string(&args.path)
            .await
            .map_err(|e| eyre::eyre!("Failed to read '{}': {}", args.path, e))?;

        if args.start_line.is_some() || args.end_line.is_some() {
            let lines: Vec<&str> = content.lines().collect();
            let start = args.start_line.unwrap_or(1).saturating_sub(1);
            let end = args.end_line.unwrap_or(lines.len()).min(lines.len());

            if start >= lines.len() {
                bail!("Start line exceeds file length");
            }

            let selected: Vec<String> = lines[start..end]
                .iter()
                .enumerate()
                .map(|(i, line)| format!("{:4} | {}", start + i + 1, line))
                .collect();

            Ok(selected.join("\n"))
        } else {
            Ok(content)
        }
    }

    fn preview(&self, args: &Value) -> String {
        if let Ok(args) = serde_json::from_value::<FsReadArgs>(args.clone()) {
            match (args.start_line, args.end_line) {
                (Some(s), Some(e)) => format!("Read {} (lines {}-{})", args.path, s, e),
                (Some(s), None) => format!("Read {} (from line {})", args.path, s),
                (None, Some(e)) => format!("Read {} (to line {})", args.path, e),
                (None, None) => format!("Read file: {}", args.path),
            }
        } else {
            "Read file (invalid args)".to_string()
        }
    }
}
