import asyncio
from typing import Dict, Any

from .base import BaseAgent
from ..workers.progress_publisher import publish_progress

class ProductManagerAgent(BaseAgent):
    """
    The Product Manager Agent validates the generated artifacts against requirements.
    """

    def __init__(self):
        super().__init__("pm")

    async def run(self, context: Dict[str, Any]) -> Dict[str, Any]:
        build_id = context["build_id"]
        self.logger.info(f"[{build_id}] Starting final validation...")
        await publish_progress(build_id, self.agent_name, "IN_PROGRESS", "Starting final validation of all generated artifacts.")

        # Simulate validation checks
        await asyncio.sleep(1)
        
        errors = []
        # Example check: Does the code contain what the blueprint specified?
        if "api/routes.py" not in context.get("code_artifacts", {}):
            errors.append("Validation failed: Backend routes file is missing.")
        
        # Example check: Were integrations configured?
        required_integrations = context.get("integration_requirements", [])
        configured_integrations = context.get("integration_details", {}).keys()
        if not set(required_integrations).issubset(set(configured_integrations)):
             errors.append("Validation failed: Not all required integrations were configured.")

        if errors:
            self.logger.error(f"[{build_id}] Validation failed with errors: {errors}")
            await publish_progress(build_id, self.agent_name, "FAILED", "Validation failed.", artifacts={"errors": errors})
            return {"validation_status": "failed", "errors": errors}

        self.logger.info(f"[{build_id}] Validation successful. Build is approved.")
        await publish_progress(build_id, self.agent_name, "COMPLETED", "All artifacts validated and approved.")

        return {"validation_status": "passed"}
