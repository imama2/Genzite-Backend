import asyncio
from typing import Dict, Any

from .base import BaseAgent
from ..workers.progress_publisher import publish_progress

class IntegrationAgent(BaseAgent):
    """
    The Integration Agent configures third-party services.
    """

    def __init__(self):
        super().__init__("integration")

    async def run(self, context: Dict[str, Any]) -> Dict[str, Any]:
        build_id = context["build_id"]
        integration_requirements = context.get("integration_requirements", [])

        if not integration_requirements:
            self.logger.info(f"[{build_id}] No integrations required. Skipping.")
            await publish_progress(build_id, self.agent_name, "COMPLETED", "No integrations required.")
            return {"integration_details": {}}

        self.logger.info(f"[{build_id}] Starting integration setup for: {integration_requirements}")
        await publish_progress(build_id, self.agent_name, "IN_PROGRESS", f"Starting integration setup for: {', '.join(integration_requirements)}")

        integration_details = {}
        for integration in integration_requirements:
            await asyncio.sleep(1)
            await publish_progress(build_id, self.agent_name, "IN_PROGRESS", f"Configuring {integration}...")
            
            prompt = f"Generate the configuration steps and environment variables needed for {integration}."
            await self._call_llm(prompt)
            
            # Simulate creating integration config
            integration_details[integration] = {
                "api_key_env": f"{integration.upper()}_API_KEY",
                "client_id_env": f"{integration.upper()}_CLIENT_ID",
                "status": "configured",
                "setup_instructions": f"1. Get API key from {integration} dashboard.\n2. Set {integration.upper()}_API_KEY in .env file."
            }
            self.logger.info(f"[{build_id}] {integration} configured.")

        self.logger.info(f"[{build_id}] All integrations completed.")
        await publish_progress(build_id, self.agent_name, "COMPLETED", "Integrations configured.", artifacts=integration_details)

        return {"integration_details": integration_details}
