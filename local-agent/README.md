# Local Agent

**An AI-powered CLI chat assistant with file system tooling, slash commands, gRPC integration, and automated document generation — built in Rust.**

Local Agent is an interactive command-line chat agent that connects to OpenAI-compatible chat APIs (like DeepSeek) and augments conversations with powerful local tool execution. It features a state-machine-driven chat loop, parallel tool execution, a slash command system, an on-demand gRPC server for programmatic file access, and automated README/document generation capabilities for code repositories.

---

## Features

- **🤖 AI-Powered Interactive Chat** — Full-duplex chat session with any OpenAI-compatible LLM API (supports streaming-disabled mode for reliability). Uses a robust state machine architecture (`PromptUser → HandleInput → HandleResponse → ValidateTools → ExecuteTools → ...`) modeled after the Amazon Q CLI.

- **🛠️ Extensible Tool System** — Three built-in file system tools exposed to the LLM via OpenAI function calling:
  - **`fs_read`** — Read file contents with optional line range selection.
  - **`fs_scan`** — List directory contents with configurable depth control.
  - **`ignore_scan`** — Scan directories respecting nested `.gitignore` rules (uses the popular `ignore` crate for correct `.gitignore` semantics).

- **⚡ Parallel Tool Execution** — Tools are executed concurrently using `tokio::spawn`, with full performance instrumentation (`tracing`-based timing logs).

- **🔐 Tool Approval Workflow** — By default, tools other than `fs_read` require user confirmation before execution. The `TRUST_ALL_TOOLS` environment variable (or `--trust-all-tools` flag) skips approval prompts.

- **🔌 On-Demand gRPC Server** — A gRPC service (`LocalAgentService`) can be spawned at runtime via the `/init` or `/doc_generation` slash command. It provides:
  - `RequestFileContent` — Read file contents remotely via the existing `FsReadTool`.
  - `HealthCheck` — Liveness probe.
  - The server is bound to `127.0.0.1:50051` (configurable via `GRPC_HOST` / `GRPC_PORT` env vars) and remains alive for the duration of the session.

- **📄 Automated Document Generation** — Slash commands `/init`, `/start`, and `/doc_generation` trigger a full pipeline:
  1. Scan the repository with `ignore_scan` (respecting `.gitignore`).
  2. Generate or improve a `README.md` by sending the project structure to an LLM.
  3. Kick off a codebase analysis via an API gateway endpoint.
  4. Optionally spawn the gRPC server for external tooling.

- **🔁 Slash Command System** — Context-aware commands parsed at runtime:
  - `/quit`, `/exit`, `/q` — Exit the chat session.
  - `/init`, `/start`, `/doc_generation [path]` — Launch document generation.
  - `/readme [path]` — Generate or improve a `README.md` for a given directory.

- **📝 Line-Input with History** — Uses `rustyline` for a polished terminal experience with readline-style editing, history, Ctrl+C/Ctrl+D handling.

- **🔧 Fully Configurable via Environment** — API endpoint, model, credentials, tool trust, logging level, gRPC host/port, and gateway endpoints are all configurable through environment variables or a `.env` file.

- **📊 Structured Logging** — All output is written to rolling daily log files (`logs/local-agent.log.YYYY-MM-DD`) using `tracing` and `tracing-appender`. The terminal only displays user-facing messages.

---

## Installation

### Prerequisites

- **Rust & Cargo** (1.70+ / edition 2021). Install from [rustup.rs](https://rustup.rs/).
- **A running `backend` repository** with protobuf definitions (see `build.rs` — proto files are expected at `../backend/shared/api/proto/v1/`).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/your-org/local-agent.git
cd local-agent

# Build release binary
cargo build --release

# The binary is located at ./target/release/local-agent
```

### Configure Environment

Copy the example environment file and fill in your credentials:

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Required: API Configuration
CHAT_API_URL=https://api.token-ai.cn/v1/chat/completions
CHAT_API_KEY=sk-your-api-key-here
CHAT_MODEL=DeepSeek-V3

# Optional: Logging level (trace/debug/info/warn/error)
RUST_LOG=info

# Optional: Tool trust
TRUST_ALL_TOOLS=false

# Optional: gRPC server
GRPC_HOST=127.0.0.1
GRPC_PORT=50051

# Optional: Gateway for codebase analysis
GATEWAY_ANALYSIS_ENDPOINT=http://localhost:8080/api/v1/analyze
```

---

## Usage

### Start an Interactive Chat Session

```bash
# Basic chat
cargo run --release -- chat

# Chat with an initial message
cargo run --release -- chat "List the files in this project"

# Trust all tools (skip approval prompts)
cargo run --release -- chat --trust-all-tools
```

Once inside the chat:

```
Welcome to Chat CLI!
Type /quit to exit

You: Hello!
Doc-agent: Hello! How can I help you today?

You: Show me the structure of the src/ directory
Doc-agent: I'll scan that for you.

Tools to execute:
  - Scan src (depth: 1)

Approve? [y/N]: y

Executing tools...
Tools executed.

Doc-agent: Here's the structure of src/:
[D] src
  [F] main.rs
  [F] lib.rs
  [F] api.rs
  ...
```

### Using Slash Commands

```
You: /readme
Found existing README.md — improving it...
✓ README.md written to ./README.md

You: /init
Starting document generation...
Please keep the agent running while documents are being generated.
✓ gRPC server started successfully on 127.0.0.1:50051
✓ Codebase analysis started successfully.

You: /quit
Goodbye!
```

### Programmatic Messaging (via API)

The `ChatSession` also supports a `send_message()` method for non-interactive usage:

```rust
use local_agent::cli::chat::{ChatArgs, ChatSession};

let args = ChatArgs::default();
let mut session = ChatSession::new(args).await?;
let response = session.send_message("What's in this project?").await?;
println!("{}", response);
```

### gRPC Service (after `/init`)

```bash
# Check health
grpcurl -plaintext 127.0.0.1:50051 api.proto.v1.LocalAgentService/HealthCheck

# Request file content
grpcurl -plaintext -d '{"args":[{"id":"1","path":"README.md"}]}' \
  127.0.0.1:50051 api.proto.v1.LocalAgentService/RequestFileContent
```

---

## Project Architecture

```
local-agent/
├── Cargo.toml                  # Dependencies: tokio, clap, tonic, reqwest, etc.
├── build.rs                    # Compiles protobuf definitions into Rust code
├── .env.example                # Environment variable template
├── src/
│   ├── main.rs                 # Entry point: env loading, logging, CLI routing
│   ├── lib.rs                  # Public module exports (api, cli, grpc, util)
│   ├── api/
│   │   └── client.rs           # ApiClient — HTTP client for LLM API calls
│   ├── cli/
│   │   ├── root.rs             # Cli struct & RootSubcommand (Chat)
│   │   └── chat/
│   │       ├── args.rs         # ChatArgs: CLI arguments & env defaults
│   │       ├── session.rs      # ChatSession: state machine engine
│   │       ├── state.rs        # ChatState enum (PromptUser, HandleInput, etc.)
│   │       ├── message.rs      # Message enum (User, Assistant, Tool)
│   │       ├── conversation.rs # ConversationHistory
│   │       ├── input_source.rs # rustyline-based user input
│   │       ├── parser.rs       # API response parser
│   │       ├── tool_manager.rs # Tool orchestration & parallel execution
│   │       ├── cli/
│   │       │   └── slash_commands/
│   │       │       ├── mod.rs  # Slash command parser/dispatcher
│   │       │       ├── quit.rs
│   │       │       ├── doc_gen.rs  # Full doc generation pipeline
│   │       │       └── readme.rs   # README generation/improvement
│   │       └── tools/
│   │           ├── registry.rs # Tool trait & ToolRegistry
│   │           ├── fs_read.rs
│   │           ├── fs_scan.rs
│   │           └── ignore_scan.rs
│   ├── grpc/
│   │   ├── server.rs           # Spawn/shutdown gRPC server (tonic)
│   │   ├── service.rs          # LocalAgentService trait implementation
│   │   └── handler.rs          # Business logic (reuses FsReadTool)
│   └── util.rs                 # Global utilities
```

---

## Contributing

Contributions are welcome! To get started:

1. **Fork** the repository.
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`).
3. **Make your changes** — the codebase follows a clean module structure. Key areas for contribution:
   - Add new tools in `src/cli/chat/tools/` by implementing the `Tool` trait.
   - Add new slash commands in `src/cli/chat/cli/slash_commands/`.
   - Improve the state machine in `session.rs`.
4. **Run the tests** — ensure `cargo build --release` and `cargo test` pass.
5. **Submit a pull request** with a clear description of your changes.

### Development Setup

```bash
# Clone with the required backend (for protobufs)
git clone https://github.com/your-org/local-agent.git
git clone https://github.com/your-org/backend.git  # sibling directory

# Build
cd local-agent
cargo build

# Run with debug logging
RUST_LOG=debug cargo run -- chat
```

---

## License

This project is licensed under the terms of the MIT license.