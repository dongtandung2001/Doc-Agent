use crate::cli::chat::tools::Tool;
use crate::cli::chat::ChatArgs;
use crate::cli::chat::ChatSession;

use super::super::super::ChatState;

use super::super::super::tools;

pub async fn execute(path: &str) -> ChatState {
    println!("Starting document generation...");
    println!("Please keep the agent running while documents are being generated.");

    // scan directory with gitignore support as part of doc generation
    let scanTool = tools::IgnoreScanTool;
    let args = serde_json::json!({ "path": path });
    let res = scanTool.invoke(args).await;

    println!(
        "Directory scan completed. Generating documents...\n{:?}",
        res
    );

    let dir = match res {
        Ok(d) => d,
        Err(e) => {
            println!("Error during directory scan: {}", e);
            return ChatState::PromptUser {
                skip_printing_tools: false,
            };
        }
    };

    let _readme = generate_readme(&dir).await;

    println!("README generation completed.\n{:?}", _readme);
    ChatState::PromptUser {
        skip_printing_tools: true,
    }
}

pub async fn generate_readme(dir: &str) -> String {
    let prompt = format!(
        "
 You are a professional code analysis expert tasked with creating a README.md document for a GitHub repository. Your goal is to analyze the content of the repository based on the provided catalogue structure and generate a high-quality README that highlights the project's key features and follows the style of advanced open-source projects on GitHub.

Here is the catalogue structure of the repository:

<catalogue>
{}
</catalogue>

The catalogue already contains the full directory and file structure of the repository. You must NOT perform any new scans, searches, or file enumerations to rediscover structure. Only use the READ_FILE function to read the contents of specific files that are already listed in the catalogue, and only when necessary to extract information for the README.

To collect information about the files in the repository, you can use the READ_FILE function. This function accepts the file path as a parameter and returns the content of the file. Use this function to read the contents of specific files mentioned in the directory. Do NOT perform any additional scans or searches beyond the provided catalogue.

Follow these steps to generate the README:

1. Essential File Analysis
   - Examine key files by using the READ_FILE function on:
     - Main project file (typically in root directory)
     - Configuration files (package.json, setup.py, Cargo.toml, etc.)
     - Documentation files (in root or /docs directory)
     - Example files or usage demonstrations
   - Only read files that are already listed in the catalogue. Do NOT trigger any repository scans.

2. Section-by-Section Information Gathering
   For each README section, READ specific files to extract accurate information:

   a. Project Title/Description
      - READ main files and configuration files
      - Look for project descriptions in package.json, setup.py, Cargo.toml, or main implementation files

   b. Features
      - READ implementation files to identify capabilities and functionality
      - Examine code structure to determine feature sets
      - Look for feature documentation in specialized files

   c. Installation
      - READ setup files like package.json, Cargo.toml, requirements.txt, or installation guides
      - Extract dependency information and setup requirements

   d. Usage
      - READ example files, documentation, or main implementation files
      - Extract code examples showing how to use the project

   e. Contributing
      - READ CONTRIBUTING.md or similar contribution guidelines

   f. License
      - READ the LICENSE file if it exists in the repository

3. README Structure
   Structure your README.md with the following sections:

   a. Project Title and Description
      - Clear, concise project name
      - Brief overview of purpose and value proposition
      - Any badges or status indicators if applicable

   b. Features
      - Bulleted list of key capabilities
      - Brief explanations of main functionality
      - What makes this project unique or valuable

   c. Installation
      - Step-by-step instructions
      - Dependencies and requirements
      - Platform-specific notes if applicable

   d. Usage
      - Basic examples with code snippets
      - Common use cases
      - API overview if applicable

   e. Contributing
      - Guidelines for contributors
      - Development setup
      - Pull request process

   f. License (ONLY if a LICENSE file exists)
      - Brief description of the license type and implications

Important Guidelines:
- The provided catalogue already contains all file and directory names. Do NOT trigger any new directory scans or searches.
- ALL information in the README MUST be obtained by READING actual file contents using the READ_FILE function.
- Only read files that are explicitly listed in the provided catalogue.
- Do NOT make assumptions about the project without verifying through file contents.
- Use Markdown formatting to enhance readability (headings, code blocks, lists, etc.).
- Focus on creating a professional, engaging README that highlights the project's strengths.
- Ensure the README is well-structured and provides clear, accurate information.

Provide your final README.md content within <readme> tags. Include no explanations or comments outside of these tags.

",
        dir
    );
    let chat_args = ChatArgs::default();

    let mut api_session = ChatSession::new(chat_args).await.unwrap();

    let response = api_session.send_message(prompt).await.unwrap();

    println!("Received response from API: {}", response);

    "".to_string()
}
