
# llm_client.py
from typing import List, Dict, Optional, TYPE_CHECKING
from openai import OpenAI

if TYPE_CHECKING:
    from openai.types.chat import ChatCompletion

from src.config import Config
from src.logger import setup_logger

logger = setup_logger(__name__)

class LLMClient:
    def __init__(self):
        if not Config.OPENAI_API_KEY:
            raise ValueError(
                "OPENAI_API_KEY is required. Please set it in your .env file or as an environment variable."
            )
        self.client = OpenAI(api_key=Config.OPENAI_API_KEY)
        self.model = Config.LLM_MODEL

    def generate_response(
            self,
            messages: List[Dict[str, str]],
            system_prompt: Optional[str] = None
    ) -> "ChatCompletion":
        """
        Generate response from LLM.
        Returns: Full OpenAI ChatCompletion response object
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

            # Call LLM and return full response
            response = self.client.chat.completions.create(
                model=self.model,
                messages=formatted_messages,
                temperature=Config.LLM_TEMPERATURE,
                max_tokens=Config.LLM_MAX_TOKENS
            )

            content_length = len(response.choices[0].message.content) if response.choices[0].message.content else 0
            logger.info(f"Generated response with {content_length} characters")
            return response

        except Exception as e:
            logger.error(f"Error generating LLM response: {str(e)}")
            raise