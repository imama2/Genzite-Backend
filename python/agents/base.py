from abc import ABC, abstractmethod
import logging
from typing import Dict, Any

class BaseAgent(ABC):
    """Abstract base class for all agents."""

    def __init__(self, agent_name: str):
        self.agent_name = agent_name
        self.logger = logging.getLogger(f"agent.{self.agent_name}")
        self.logger.setLevel(logging.INFO)

    @abstractmethod
    async def run(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        The main entry point for an agent to perform its task.

        Args:
            context: A dictionary containing the current state of the build,
                     including build_id, user_prompt, and artifacts from
                     previous agents.

        Returns:
            A dictionary containing the results of the agent's work.
            This will be merged back into the main context.
        """
        pass

    async def _call_llm(self, prompt: str, temperature: float = 0.1) -> str:
        """
        A placeholder for making a call to a large language model.
        In a real implementation, this would use a library like langchain
        or a direct API call to OpenAI, Anthropic, etc.
        """
        self.logger.info(f"Simulating LLM call for prompt: {prompt[:100]}...")
        # In a real scenario, you would have something like:
        # from langchain.llms import OpenAI
        # llm = OpenAI(temperature=temperature, api_key=settings.OPENAI_API_KEY)
        # response = await llm.apredict(prompt)
        import asyncio
        await asyncio.sleep(1)  # Simulate network latency
        self.logger.info("LLM call simulation complete.")
        return f"LLM response for: {prompt[:50]}"
