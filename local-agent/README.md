# Local Agent

An AI-powered chat assistant with tools, built in Rust.

## Features

- **Interactive Chat**: Start an interactive chat session with the AI assistant.
- **Tool Integration**: Supports tools like file scanning (`fs_scan`, `fs_read`, `ignore_scan`) for enhanced functionality.
- **Environment Support**: Loads `.env` files for configuration, including API keys and model settings.
- **Logging**: Built-in logging with `tracing` for debugging and monitoring.
- **CLI Interface**: Command-line interface powered by `clap` for easy interaction.
- **Async Runtime**: Utilizes `tokio` for efficient asynchronous operations.
- **Customizable**: Configure API endpoints, models, and tool behavior via environment variables.

## Installation

1. Ensure you have Rust and Cargo installed. If not, install them from [rustup.rs](https://rustup.rs/).
2. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/local-agent.git
   cd local-agent
   ```
3. Build the project:
   ```bash
   cargo build --release
   ```
4. Run the executable:
   ```bash
   cargo run --release
   ```

## Usage

### Starting an Interactive Chat Session

To start an interactive chat session, run:
```bash
local-agent chat
```

### Environment Configuration

The application supports loading environment variables from a `.env` file. Place your `.env` file in the root directory of the project. Example `.env` file:
```env
CHAT_API_URL=https://api.token-ai.cn/v1/chat/completions
CHAT_API_KEY=your_api_key
CHAT_MODEL=DeepSeek-V3
TRUST_ALL_TOOLS=false
```

### Tool Usage

The chat session supports tools for file operations:
- **File Scanning**: Use `fs_scan` to list directory contents.
- **File Reading**: Use `fs_read` to read file contents.
- **Gitignore Support**: Use `ignore_scan` to scan directories while respecting `.gitignore` rules.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a clear description of your changes.

## License

This project is licensed under the terms of the MIT license. See the `LICENSE` file for details.