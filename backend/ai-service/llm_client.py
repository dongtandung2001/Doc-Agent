
# llm_client.py
from typing import List, Dict, Optional, Tuple
from openai import OpenAI
from config import Config
from logger import setup_logger

logger = setup_logger(__name__)

class LLMClient:
    def __init__(self):
        self.client = OpenAI(api_key=Config.OPENAI_API_KEY)
        self.model = Config.LLM_MODEL

    def generate_response(
            self,
            messages: List[Dict[str, str]],
            system_prompt: Optional[str] = None
    ) -> Tuple[str, bool]:
        """
        Generate response from LLM.
        Returns: (response_content, tools_executed)
        """
        try:
            # Prepare messages
            formatted_messages = []

            if system_prompt:
                formatted_messages.append({
                    "role": "system",
                    "content": system_prompt
                })

            formatted_messages.extend(messages)

            # Call LLM
            response = self.client.chat.completions.create(
                model=self.model,
                messages=formatted_messages,
                temperature=Config.LLM_TEMPERATURE,
                max_tokens=Config.LLM_MAX_TOKENS
            )

            # Extract response
            content = response.choices[0].message.content

            # Check if tools were executed (simplified - extend as needed)
            tools_executed = False
            if hasattr(response.choices[0].message, 'tool_calls'):
                tools_executed = response.choices[0].message.tool_calls is not None

            logger.info(f"Generated response with {len(content)} characters")
            return content, tools_executed

        except Exception as e:
            logger.error(f"Error generating LLM response: {str(e)}")
            raise