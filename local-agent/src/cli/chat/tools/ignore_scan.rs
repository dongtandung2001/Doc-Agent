use async_trait::async_trait;
use eyre::Result;
use ignore::WalkBuilder;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

use super::registry::Tool;

pub struct IgnoreScanTool;

#[derive(Debug, Deserialize, Serialize)]
struct IgnoreScanArgs {
    path: String,
    #[serde(default)]
    max_depth: Option<usize>,
}

#[async_trait]
impl Tool for IgnoreScanTool {
    fn name(&self) -> &str {
        "ignore_scan"
    }

    fn description(&self) -> &str {
        "Scan directory with gitignore support (respects nested .gitignore files)"
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
                    "description": "Max depth (omit or null for unlimited depth)"
                }
            },
            "required": ["path"]
        })
    }

    async fn invoke(&self, args: Value) -> Result<String> {
        let args: IgnoreScanArgs = serde_json::from_value(args)?;
        let result = scan_with_gitignore(&args.path, args.max_depth).await?;
        Ok(result.join("\n"))
    }

    fn preview(&self, args: &Value) -> String {
        if let Ok(args) = serde_json::from_value::<IgnoreScanArgs>(args.clone()) {
            let depth_str = args
                .max_depth
                .map(|d| d.to_string())
                .unwrap_or_else(|| "unlimited".to_string());
            format!("Scan {} with gitignore (depth: {})", args.path, depth_str)
        } else {
            "Scan directory with gitignore (invalid args)".to_string()
        }
    }
}

/// Scan directory with gitignore support (handles nested .gitignore files)
///
/// This function uses the `ignore` crate to properly handle:
/// - Nested .gitignore files
/// - Negation rules (!pattern)
/// - Global gitignore
/// - .git/info/exclude
/// - Correct pattern matching
///
/// # Arguments
/// * `path` - Directory path to scan
/// * `max_depth` - Maximum depth to scan. Use `None` for unlimited depth, or `Some(depth)` to limit
async fn scan_with_gitignore(path: &str, max_depth: Option<usize>) -> Result<Vec<String>> {
    // Run the synchronous WalkBuilder in a blocking task
    let path_owned = path.to_string();
    let entries = tokio::task::spawn_blocking(move || {
        // Convert max_depth for WalkBuilder (which counts root as depth 0)
        let walker_depth = max_depth.map(|d| d + 1);

        let walker = WalkBuilder::new(&path_owned)
            .max_depth(walker_depth) // None = unlimited, Some(n) = limited
            .git_ignore(true) // Respect .gitignore files (including nested ones)
            .git_global(true) // Respect global gitignore
            .git_exclude(true) // Respect .git/info/exclude
            .hidden(false) // Don't automatically skip hidden files
            .parents(true) // Check parent directories for .gitignore
            .build();

        let mut collected = Vec::new();
        for entry in walker.filter_map(|e| e.ok()) {
            let entry_path = entry.path();
            let depth = entry.depth();

            // Skip the root directory itself
            if depth == 0 {
                continue;
            }

            let indent = "  ".repeat(depth - 1);
            let icon = if entry.file_type().map_or(false, |ft| ft.is_dir()) {
                "[D]"
            } else {
                "[F]"
            };

            let name = entry_path.file_name().unwrap_or_default().to_string_lossy();

            collected.push(format!("{}{} {}", indent, icon, name));
        }
        collected
    })
    .await?;

    println!(
        "Scanned directory with gitignore: {} entries",
        entries.len()
    );

    Ok(entries)
}
