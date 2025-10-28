use eyre::Result;
use rustyline::error::ReadlineError;
use rustyline::DefaultEditor;

/// Wraps rustyline for user input
pub struct InputSource {
    editor: DefaultEditor,
}

impl InputSource {
    pub fn new() -> Result<Self> {
        let editor = DefaultEditor::new()?;
        Ok(Self { editor })
    }

    /// Read a line of input with the given prompt
    pub fn read_line(&mut self, prompt: &str) -> Result<Option<String>> {
        match self.editor.readline(prompt) {
            Ok(line) => {
                let trimmed = line.trim();
                if !trimmed.is_empty() {
                    self.editor.add_history_entry(&line)?;
                }
                Ok(Some(line))
            }
            Err(ReadlineError::Interrupted) => {
                // Ctrl+C
                Ok(None)
            }
            Err(ReadlineError::Eof) => {
                // Ctrl+D
                Ok(None)
            }
            Err(err) => Err(err.into()),
        }
    }
}
