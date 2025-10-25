use anyhow::{Context, Result};
use clap::Parser;
use std::io::{self, Write};
use std::path::{Path, PathBuf};

/// Tiny CLI: read a path and print info (no backend)
#[derive(Parser, Debug)]
#[command(name = "path-printer", version, about = "Print info about a local path")]
struct Args {
    /// Path to a file or directory (if omitted, we will prompt)
    #[arg(long)]
    path: Option<PathBuf>,
}

fn main() -> Result<()> {
    let args = Args::parse();

    // 1) get path from args or prompt
    let path = match args.path {
        Some(p) => p,
        None => prompt_for_path("Enter a path: ")?,
    };

    // 2) basic validation
    ensure_readable(&path).with_context(|| format!("Path not usable: {}", path.display()))?;

    // 3) print canonical absolute path + type
    let canon = path.canonicalize().context("Failed to canonicalize path")?;
    let meta = std::fs::metadata(&canon)?;
    let kind = if meta.is_dir() {
        "directory"
    } else if meta.is_file() {
        "file"
    } else {
        "other"
    };

    println!("✅ OK");
    println!("  Absolute: {}", canon.display());
    println!("  Kind:     {}", kind);

    Ok(())
}

fn prompt_for_path(msg: &str) -> Result<PathBuf> {
    print!("{msg}");
    io::stdout().flush().ok();
    let mut s = String::new();
    io::stdin().read_line(&mut s).context("Failed to read input")?;
    let trimmed = s.trim();
    if trimmed.is_empty() {
        anyhow::bail!("No path provided");
    }
    Ok(PathBuf::from(trimmed))
}

fn ensure_readable(p: &Path) -> Result<()> {
    if !p.exists() {
        anyhow::bail!("Does not exist: {}", p.display());
    }
    let meta = std::fs::metadata(p)?;
    if meta.is_file() {
        // try opening to confirm readability
        std::fs::File::open(p).context("Cannot open file for reading")?;
    }
    Ok(())
}
