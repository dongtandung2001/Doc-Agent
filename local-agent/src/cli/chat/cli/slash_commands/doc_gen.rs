use super::super::super::ChatState;

pub fn execute() -> ChatState {
    println!("Starting document generation...");
    println!("Please keep the agent running while documents are being generated.");
    ChatState::PromptUser {
        skip_printing_tools: true,
    }
}
