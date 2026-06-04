import logging
import json
from typing import AsyncIterator

from .generated import architect_pb2, architect_pb2_grpc
from agents.architect import ArchitectAgent

logger = logging.getLogger(__name__)

class ArchitectServicer(architect_pb2_grpc.ArchitectServicer):
    """
    Provides the gRPC service for the Architect Agent.
    """

    def __init__(self):
        self.agent = ArchitectAgent()

    async def AnalyzePrompt(
        self,
        request_iterator: AsyncIterator[architect_pb2.AnalyzeRequest],
        context,
    ) -> AsyncIterator[architect_pb2.AnalyzeResponse]:
        """
        Handles the bidirectional streaming RPC for prompt analysis.
        """
        logger.info("AnalyzePrompt stream started.")
        
        # For this placeholder, we'll just process the first message and run our logic.
        # A real implementation would handle a full conversation.
        first_request = await anext(request_iterator)
        
        # Extract the initial prompt from the oneof field
        if first_request.HasField('initial_prompt'):
            initial = first_request.initial_prompt
            build_id = initial.build_id
            user_input = initial.user_prompt
            history = []
        else:
            logger.error("First request must contain initial_prompt")
            return

        try:
            async for result in self.agent.analyze_prompt_stream(build_id, user_input, history):
                status = result.get("status", "IN_PROGRESS")
                
                # Create response based on status
                if status == "ASKING_QUESTION":
                    question = architect_pb2.AnalyzeResponse.ClarificationQuestion(
                        question_id=result.get("question_id", ""),
                        text=result.get("message", ""),
                        options=result.get("options", [])
                    )
                    response = architect_pb2.AnalyzeResponse(question=question)
                elif status == "COMPLETED":
                    update = architect_pb2.AnalyzeResponse.BlueprintUpdate(
                        blueprint_chunk_json=result.get("blueprint_json", "")
                    )
                    response = architect_pb2.AnalyzeResponse(update=update)
                    yield response
                    # Signal conversation complete
                    complete = architect_pb2.AnalyzeResponse.ConversationComplete()
                    response = architect_pb2.AnalyzeResponse(conversation_complete=complete)
                else:  # IN_PROGRESS
                    update = architect_pb2.AnalyzeResponse.BlueprintUpdate(
                        blueprint_chunk_json=result.get("blueprint_json", "")
                    )
                    response = architect_pb2.AnalyzeResponse(update=update)
                
                yield response
        except Exception as e:
            logger.error(f"Error during prompt analysis for build {build_id}: {e}", exc_info=True)
            # Send error as a blueprint update with error information
            update = architect_pb2.AnalyzeResponse.BlueprintUpdate(
                blueprint_chunk_json=json.dumps({"error": str(e)})
            )
            yield architect_pb2.AnalyzeResponse(update=update)
        
        logger.info(f"AnalyzePrompt stream for build {build_id} finished.")

    async def FinalizeBlueprint(self, request: architect_pb2.FinalizeBlueprintRequest, context):
        """
        Handles the unary RPC for finalizing a blueprint.
        (Placeholder implementation)
        """
        logger.info(f"FinalizeBlueprint called for build {request.build_id}")
        # In a real implementation, this might trigger a final LLM call
        # based on the user's confirmation.
        
        # For now, we'll just return a success response with a dummy blueprint.
        from agents.architect import Blueprint, DataModel, Endpoint, Page
        dummy_blueprint = Blueprint(
            data_models=[DataModel(name="User", fields={"email": "string"})],
            api_endpoints=[Endpoint(path="/api/v1/auth/login", method="POST", description="User login")],
            pages=[Page(name="Homepage", path="/", components=["Header", "Hero"])]
        )
        
        return architect_pb2.FinalizeBlueprintResponse(
            blueprint_json=dummy_blueprint.model_dump_json(indent=2)
        )

# Helper to get the first item from an async iterator
async def anext(ait):
    return await ait.__anext__()
