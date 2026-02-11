import os
from openai import OpenAI
from traceloop.sdk import Traceloop


class SimpleAgent:
    def __init__(self):
        # Initialize Traceloop for tracing
        # Read env variables first
        api_endpoint = os.getenv("AMP_OTEL_ENDPOINT")
        api_key = os.getenv("AMP_AGENT_API_KEY")
        
        # Print them BEFORE init
        print(f"AMP_OTEL_ENDPOINT: {api_endpoint}")
        print(f"AMP_AGENT_API_KEY: {api_key}")
        
        # Then initialize
        Traceloop.init(
            app_name="simple-openai-agent",
            api_endpoint=api_endpoint,
            disable_batch=True,
            headers={"x-amp-api-key": api_key},
        )

        # Initialize OpenAI client
        self.client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))
        self.model = os.getenv("OPENAI_MODEL", "gpt-4o-mini")

    def chat(self, message: str, session_id: str, context: dict = None) -> str:
        """
        Send a message to OpenAI and get a response.

        Args:
            message: User message to send
            session_id: Session identifier for tracking
            context: Optional context as a dictionary

        Returns:
            Assistant's response
        """
        try:
            # Build system prompt with context if provided
            system_prompt = "You are a helpful AI assistant."
            if context:
                context_str = "\n".join([f"{k}: {v}" for k, v in context.items()])
                system_prompt += f"\n\nContext:\n{context_str}"

            response = self.client.chat.completions.create(
                model=self.model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": message},
                ],
                temperature=0.7,
                max_tokens=1000,
                user=session_id,  # Track user/session in OpenAI
            )

            return response.choices[0].message.content
        except Exception as e:
            return f"Error: {str(e)}"


if __name__ == "__main__":
    # Example usage
    agent = SimpleAgent()

    print("Testing agent:")
    response = agent.chat(
        message="What is the capital of France?",
        session_id="test-session-123"
    )
    print(f"Response: {response}")
