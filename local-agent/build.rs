// local-agent/build.rs
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // <crate root>/local-agent  -> go up one level -> <repo root>
    let repo_root = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .expect("local-agent has no parent")
        .to_path_buf();

    let local_agent_dir = repo_root.join("local-agent");
    let grpc_dir = local_agent_dir.join("src").join("grpc");

    let backend_dir = repo_root.join("backend");
    let include_dir = backend_dir.join("shared"); // contains "api/..."

    // Explicit paths to the protos
    let protos = [
        backend_dir.join("shared/api/proto/v1/localagent.proto"),
        backend_dir.join("shared/api/proto/v1/common.proto"),
    ];

    // Re-run build if these change
    println!("cargo:rerun-if-changed={}", include_dir.display());
    for p in &protos {
        println!("cargo:rerun-if-changed={}", p.display());
    }

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(true)
        .compile_protos(&protos, &[include_dir])?;

    Ok(())
}
