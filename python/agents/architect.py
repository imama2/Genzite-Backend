import asyncio
import json
from typing import AsyncIterator, List, Dict, Any

from .base import BaseAgent

# Pydantic models for structured data
from pydantic import BaseModel, Field

class DataModel(BaseModel):
    name: str
    fields: Dict[str, str]

class Endpoint(BaseModel):
    path: str
    method: str
    description: str

class Page(BaseModel):
    name: str
    path: str
    components: List[str]

class Blueprint(BaseModel):
    data_models: List[DataModel] = Field(default_factory=list)
    api_endpoints: List[Endpoint] = Field(default_factory=list)
    pages: List[Page] = Field(default_factory=list)


class ArchitectAgent(BaseAgent):
    """
    The Architect Agent interprets the user's prompt into a technical blueprint.
    Its logic is primarily exposed via a gRPC service for interactive clarification.
    """

    def __init__(self):
        super().__init__("architect")

    async def analyze_prompt_stream(
        self, build_id: str, user_input: str, conversation_history: List[Dict[str, str]]
    ) -> AsyncIterator[Dict[str, Any]]:
        """
        Analyzes the user prompt and yields clarifying questions or a final blueprint.
        This is the core logic called by the gRPC `AnalyzePrompt` method.
        """
        self.logger.info(f"[{build_id}] Starting analysis of prompt: '{user_input[:50]}...'")

        # 1. Initial analysis (simulated LLM call)
        yield {
            "status": "IN_PROGRESS",
            "message": "Analyzing initial prompt to identify core requirements...",
        }
        analysis_prompt = f"Analyze the following user prompt for a website and identify ambiguities. User prompt: {user_input}"
        initial_analysis = await self._call_llm(analysis_prompt)
        self.logger.info(f"[{build_id}] Initial analysis result: {initial_analysis}")
        await asyncio.sleep(1)

        # 2. Ask a clarifying question (placeholder)
        yield {
            "status": "ASKING_QUESTION",
            "message": "What kind of user authentication is needed? (e.g., none, email/password, social login)",
        }
        # In a real scenario, you would wait for the user's response here.
        # The gRPC stream would handle the back-and-forth.
        await asyncio.sleep(1) # Simulate waiting for an answer

        # 3. Ask another question
        yield {
            "status": "IN_PROGRESS",
            "message": "Processing response... now checking for data requirements.",
        }
        await asyncio.sleep(1)
        yield {
            "status": "ASKING_QUESTION",
            "message": "Should there be a contact form? If so, what fields should it have?",
        }
        await asyncio.sleep(1)

        # 4. Finalize blueprint
        yield {
            "status": "IN_PROGRESS",
            "message": "All questions answered. Generating final blueprint...",
        }
        blueprint_prompt = f"Based on the prompt '{user_input}' and the conversation, generate a technical blueprint."
        final_blueprint_str = await self._call_llm(blueprint_prompt)
        
        # Create a dummy blueprint for now
        dummy_blueprint = Blueprint(
            data_models=[DataModel(name="User", fields={"email": "string", "password_hash": "string"})],
            api_endpoints=[Endpoint(path="/api/v1/auth/login", method="POST", description="User login")],
            pages=[Page(name="Homepage", path="/", components=["Header", "Hero", "Footer"])]
        )
        
        self.logger.info(f"[{build_id}] Final blueprint generated.")
        yield {
            "status": "COMPLETED",
            "message": "Blueprint finalized.",
            "blueprint_json": dummy_blueprint.model_dump_json(indent=2),
        }

    async def run(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        The Architect agent does not run in the main build worker pipeline.
        Its work is done upfront via the gRPC server.
        This method is a no-op and should not be called.
        """
        self.logger.warning("ArchitectAgent.run() was called, but it should not be part of the async worker pipeline.")
        return {}
