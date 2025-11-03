use eyre::Result;
use serde_json::json;
use tokio::sync::broadcast;
use tracing::error;

use super::args::ChatArgs;
use super::cli::slash_commands;
use super::conversation::ConversationHistory;
use super::input_source::InputSource;
use super::parser::parse_response;
use super::state::{ChatState, ToolUse};
use super::tool_manager::ToolManager;
use crate::api::ApiClient;

pub struct ChatSession {
    /// Current state
    inner: Option<ChatState>,

    /// Conversation history
    conversation: ConversationHistory,

    /// User input
    input_source: InputSource,

    /// Tool manager
    tool_manager: ToolManager,

    /// API client
    api_client: ApiClient,

    /// Ctrl+C channel
    ctrlc_rx: broadcast::Receiver<()>,
}

impl ChatSession {
    pub async fn new(args: ChatArgs) -> Result<Self> {
        // Debug: print the loaded configuration
        eprintln!("Debug - API Configuration:");
        eprintln!("  URL: {}", args.api_url);
        eprintln!(
            "  API Key: {}***",
            if args.api_key.len() > 10 {
                &args.api_key[..10]
            } else {
                "EMPTY"
            }
        );
        eprintln!("  Model: {}", args.model);

        // Set up Ctrl+C handler
        let (ctrlc_tx, ctrlc_rx) = broadcast::channel(4);
        tokio::spawn(async move {
            loop {
                match tokio::signal::ctrl_c().await {
                    Ok(_) => {
                        let _ = ctrlc_tx.send(());
                    }
                    Err(err) => {
                        error!(?err, "Ctrl+C error");
                    }
                }
            }
        });

        Ok(Self {
            inner: Some(ChatState::default()),
            conversation: ConversationHistory::new(),
            input_source: InputSource::new()?,
            tool_manager: ToolManager::new(args.trust_all_tools),
            api_client: ApiClient::new(&args.api_url, &args.api_key, &args.model)?,
            ctrlc_rx,
        })
    }

    /// Main chat loop
    pub async fn spawn(&mut self) -> Result<()> {
        println!("Welcome to Chat CLI!");
        println!("Type /quit to exit\n");

        // Main loop - matches Q CLI pattern
        while !matches!(self.inner, Some(ChatState::Exit)) {
            self.next().await?;
        }

        println!("\nGoodbye!");
        Ok(())
    }

    /// Context + available tools
    /// Write, Read tool, scan, search,....
    /// Prompt + Catalog + File Structure

    /// State transition engine
    /// Large Language Model (LLM)
    /// pre-trained LLM model --> extra training on docs generation
    /// State Machine
    async fn next(&mut self) -> Result<()> {
        let mut ctrl_c_stream = self.ctrlc_rx.resubscribe();

        // Take current state
        let state = self.inner.take().expect("state must be Some");

        // Execute handler for current state
        let result = match state {
            ChatState::PromptUser {
                skip_printing_tools,
            } => {
                eprintln!("[DEBUG] State: PromptUser");
                self.prompt_user(skip_printing_tools).await
            }

            ChatState::HandleInput { input } => {
                eprintln!("[DEBUG] State: HandleInput");
                tokio::select! {
                    res = self.handle_input(input) => res,
                    Ok(_) = ctrl_c_stream.recv() => {
                        println!("\n(Press Ctrl+C again to exit)");
                        Ok(ChatState::PromptUser { skip_printing_tools: false })
                    }
                }
            }

            ChatState::HandleResponse { request } => {
                eprintln!("[DEBUG] State: HandleResponse");
                tokio::select! {
                    res = self.handle_response(request) => res,
                    Ok(_) = ctrl_c_stream.recv() => {
                        println!("\n\nInterrupted!");
                        Ok(ChatState::PromptUser { skip_printing_tools: false })
                    }
                }
            }

            ChatState::ValidateTools { tools } => {
                eprintln!("[DEBUG] State: ValidateTools");
                self.validate_tools(tools).await
            }

            ChatState::ExecuteTools { tools } => {
                eprintln!("[DEBUG] State: ExecuteTools");
                tokio::select! {
                    res = self.execute_tools(tools) => res,
                    Ok(_) = ctrl_c_stream.recv() => {
                        println!("\n\nTool execution interrupted!");
                        Ok(ChatState::PromptUser { skip_printing_tools: false })
                    }
                }
            }

            ChatState::Exit => {
                eprintln!("[DEBUG] State: Exit");
                return Ok(());
            }
        };

        // Transition to next state
        match result {
            Ok(next_state) => {
                self.inner = Some(next_state);
                Ok(())
            }
            Err(err) => {
                error!(?err, "State transition error");
                eprintln!("Error: {}", err);
                self.inner = Some(ChatState::PromptUser {
                    skip_printing_tools: false,
                });
                Ok(())
            }
        }
    }

    // State handlers will be implemented next
    async fn prompt_user(&mut self, _skip_printing_tools: bool) -> Result<ChatState> {
        match self.input_source.read_line("You: ")? {
            Some(input) => Ok(ChatState::HandleInput { input }),
            None => Ok(ChatState::Exit),
        }
    }

    async fn handle_input(&mut self, input: String) -> Result<ChatState> {
        // Check for slash commands
        if let Some(state) = slash_commands::parse_and_execute(&input) {
            return Ok(state);
        }

        // Add user message to conversation
        self.conversation
            .add(super::message::Message::user(input.clone()));

        // Transition to handling response
        Ok(ChatState::HandleResponse { request: input })
    }

    async fn handle_response(&mut self, _request: String) -> Result<ChatState> {
        print!("Doc-agent: ");

        // Convert messages to API format
        let messages: Vec<_> = self
            .conversation
            .messages()
            .iter()
            .map(|m| serde_json::to_value(m).unwrap())
            .collect();

        let tools = self.tool_manager.tool_specs();

        // eprintln!("[DEBUG] Sending request to API...");
        // eprintln!("[DEBUG] Messages count: {}", messages.len());
        // eprintln!("[DEBUG] Tools count: {}", tools.len());

        // Get JSON response from API client
        let response = self.api_client.send_message(messages, tools).await?;

        // Parse response using parser module
        let parsed = parse_response(&response)?;

        // Display assistant text if present
        if !parsed.assistant_text.is_empty() {
            println!("{}", parsed.assistant_text);
        }

        println!(); // Final newline

        // Add assistant message if there was text
        if !parsed.assistant_text.is_empty() {
            self.conversation
                .add(super::message::Message::assistant(parsed.assistant_text));
        }

        // If tools were requested, validate and execute them
        if !parsed.tools_to_execute.is_empty() {
            eprintln!(
                "[DEBUG] Total tools to execute: {}",
                parsed.tools_to_execute.len()
            );
            Ok(ChatState::ValidateTools {
                tools: parsed.tools_to_execute,
            })
        } else {
            Ok(ChatState::PromptUser {
                skip_printing_tools: false,
            })
        }
    }

    async fn validate_tools(&mut self, tools: Vec<ToolUse>) -> Result<ChatState> {
        // Show what tools will be executed
        let previews = self.tool_manager.preview_tools(&tools);
        println!("\nTools to execute:");
        for preview in &previews {
            println!("  - {}", preview);
        }

        // Check if approval is needed
        if self.tool_manager.requires_approval(&tools) {
            println!("\nApprove? [y/N]: ");
            if let Some(response) = self.input_source.read_line("")? {
                if !response.trim().eq_ignore_ascii_case("y") {
                    println!("Tools rejected.");
                    return Ok(ChatState::PromptUser {
                        skip_printing_tools: false,
                    });
                }
            } else {
                return Ok(ChatState::Exit);
            }
        }

        Ok(ChatState::ExecuteTools { tools })
    }

    async fn execute_tools(&mut self, tools: Vec<ToolUse>) -> Result<ChatState> {
        println!("\nExecuting tools...");

        // Execute tools and get results
        let results = self.tool_manager.execute_all(tools).await;

        // Add tool results to conversation
        for result in results {
            self.conversation.add(result);
        }

        println!("Tools executed.\n");

        // Continue the conversation with tool results
        Ok(ChatState::HandleResponse {
            request: String::new(),
        })
    }
}
