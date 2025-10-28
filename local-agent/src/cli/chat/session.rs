use eyre::Result;
use futures::StreamExt;
use serde_json::json;
use tokio::sync::broadcast;
use tracing::{debug, error};

use super::args::ChatArgs;
use super::cli::SlashCommand;
use super::conversation::ConversationHistory;
use super::input_source::InputSource;
use super::parser::{parse_sse_stream, StreamEvent};
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
            api_client: ApiClient::new(&args.api_url)?,
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
            } => self.prompt_user(skip_printing_tools).await,

            ChatState::HandleInput { input } => {
                tokio::select! {
                    res = self.handle_input(input) => res,
                    Ok(_) = ctrl_c_stream.recv() => {
                        println!("\n(Press Ctrl+C again to exit)");
                        Ok(ChatState::PromptUser { skip_printing_tools: false })
                    }
                }
            }

            ChatState::HandleResponseStream { request } => {
                tokio::select! {
                    res = self.handle_response(request) => res,
                    Ok(_) = ctrl_c_stream.recv() => {
                        println!("\n\nInterrupted!");
                        Ok(ChatState::PromptUser { skip_printing_tools: false })
                    }
                }
            }

            ChatState::ValidateTools { tools } => self.validate_tools(tools).await,

            ChatState::ExecuteTools { tools } => {
                tokio::select! {
                    res = self.execute_tools(tools) => res,
                    Ok(_) = ctrl_c_stream.recv() => {
                        println!("\n\nTool execution interrupted!");
                        Ok(ChatState::PromptUser { skip_printing_tools: false })
                    }
                }
            }

            ChatState::Exit => return Ok(()),
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
        if let Some(cmd) = SlashCommand::parse(&input) {
            return Ok(cmd.execute());
        }

        // Add user message to conversation
        self.conversation
            .add(super::message::Message::user(input.clone()));

        // Transition to streaming response
        Ok(ChatState::HandleResponseStream { request: input })
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

        // Get raw response from API client
        let response = self.api_client.send_message(messages, tools).await?;

        // Parse the SSE stream
        let stream = parse_sse_stream(response);
        futures::pin_mut!(stream);

        let mut assistant_text = String::new();
        let mut tools_to_execute = Vec::new();
        let mut current_tool: Option<ToolUse> = None;
        let mut tool_json_buffer = String::new();

        while let Some(event_result) = stream.next().await {
            match event_result? {
                StreamEvent::ContentDelta { text } => {
                    print!("{}", text);
                    assistant_text.push_str(&text);
                }
                StreamEvent::ToolUseStart { id, name } => {
                    if !assistant_text.is_empty() {
                        println!(); // New line after text
                    }
                    println!("\n[Tool: {}]", name);
                    current_tool = Some(ToolUse {
                        id,
                        name,
                        args: json!({}),
                    });
                    tool_json_buffer.clear();
                }
                StreamEvent::ToolInputDelta { json } => {
                    tool_json_buffer.push_str(&json);
                }
                StreamEvent::Done => {
                    break;
                }
            }
        }

        // Finalize any pending tool
        if let Some(mut tool) = current_tool.take() {
            if !tool_json_buffer.is_empty() {
                tool.args = serde_json::from_str(&tool_json_buffer)?;
            }
            tools_to_execute.push(tool);
        }

        println!(); // Final newline

        // Add assistant message if there was text
        if !assistant_text.is_empty() {
            self.conversation
                .add(super::message::Message::assistant(assistant_text));
        }

        // If tools were requested, validate and execute them
        if !tools_to_execute.is_empty() {
            Ok(ChatState::ValidateTools {
                tools: tools_to_execute,
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
            print!("\nApprove? [y/N]: ");
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

        // Add tool use messages to conversation
        for tool in &tools {
            self.conversation.add(super::message::Message::tool_use(
                &tool.id,
                &tool.name,
                tool.args.clone(),
            ));
        }

        // Execute tools
        let results = self.tool_manager.execute_all(tools).await;

        // Add results to conversation
        for result in results {
            self.conversation.add(result);
        }

        println!("Tools executed.\n");

        // Continue the conversation with tool results
        Ok(ChatState::HandleResponseStream {
            request: String::new(),
        })
    }
}
