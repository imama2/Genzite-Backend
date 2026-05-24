import logging
from typing import AsyncIterator

from .generated import architect_pb2, architect_pb2_grpc
from ..agents.architect import ArchitectAgent

logger = logging.getLogger(__name__)

class ArchitectAgentServicer(architect_pb2_grpc.ArchitectAgentServicer):
    """
    Provides the gRPC service for the Architect Agent.
    """

    def __init__(self):
        self.agent = ArchitectAgent()

    async def AnalyzePrompt(
        self,
        request_iterator: AsyncIterator[architect_pb2.PromptRequest],
        context,
    ) -> AsyncIterator[architect_pb2.PromptResponse]:
        """
        Handles the bidirectional streaming RPC for prompt analysis.
        """
        logger.info("AnalyzePrompt stream started.")
        
        # For this placeholder, we'll just process the first message and run our logic.
        # A real implementation would handle a full conversation.
        first_request = await anext(request_iterator)
        build_id = first_request.build_id
        user_input = first_request.user_input
        history = [{"role": msg.role, "content": msg.content} for msg in first_request.history]

        try:
            async for result in self.agent.analyze_prompt_stream(build_id, user_input, history):
                status_enum = architect_pb2.PromptResponse.Status.Value(result["status"])
                
                response = architect_pb2.PromptResponse(
                    build_id=build_id,
                    status=status_enum,
                    message=result.get("message", ""),
                    blueprint_json=result.get("blueprint_json", ""),
                )
                yield response
        except Exception as e:
            logger.error(f"Error during prompt analysis for build {build_id}: {e}", exc_info=True)
            yield architect_pb2.PromptResponse(
                build_id=build_id,
                status=architect_pb2.PromptResponse.Status.Value("ERROR"),
                message=f"An internal error occurred: {e}",
            )
        
        logger.info(f"AnalyzePrompt stream for build {build_id} finished.")

    async def FinalizeBlueprint(self, request: architect_pb2.FinalizeRequest, context):
        """
        Handles the unary RPC for finalizing a blueprint.
        (Placeholder implementation)
        """
        logger.info(f"FinalizeBlueprint called for build {request.build_id}")
        # In a real implementation, this might trigger a final LLM call
        # based on the user's confirmation.
        
        # For now, we'll just return a success response with a dummy blueprint.
        from ..agents.architect import Blueprint, DataModel, Endpoint, Page
        dummy_blueprint = Blueprint(
            data_models=[DataModel(name="User", fields={"email": "string"})],
            api_endpoints=[Endpoint(path="/api/v1/auth/login", method="POST", description="User login")],
            pages=[Page(name="Homepage", path="/", components=["Header", "Hero"])]
        )
        
        return architect_pb2.FinalizeResponse(
            success=True,
            blueprint_json=dummy_blueprint.model_dump_json(indent=2)
        )

# Helper to get the first item from an async iterator
async def anext(ait):
    return await ait.__anext__()
