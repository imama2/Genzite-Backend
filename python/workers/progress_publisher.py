import json
import redis.asyncio as redis
from datetime import datetime, timezone
from typing import Dict, Any, Optional

from ..config import settings

# Global redis client
_redis_client = None

async def get_redis_client():
    """Initializes and returns the Redis client."""
    global _redis_client
    if _redis_client is None:
        _redis_client = redis.Redis(
            host=settings.REDIS_HOST,
            port=settings.REDIS_PORT,
            password=settings.REDIS_PASSWORD,
            db=settings.REDIS_DB,
            decode_responses=True
        )
    return _redis_client

async def publish_progress(
    build_id: str,
    agent: str,
    status: str,
    message: str,
    artifacts: Optional[Dict[str, Any]] = None,
):
    """
    Publishes a progress event to the build-specific Redis channel.

    Args:
        build_id: The unique identifier for the build.
        agent: The name of the agent publishing the event.
        status: The status of the operation (e.g., "IN_PROGRESS", "COMPLETED", "FAILED").
        message: A human-readable message describing the progress.
        artifacts: An optional dictionary of resulting artifacts.
    """
    redis_client = await get_redis_client()
    channel = f"build:{build_id}:events"
    
    event_data = {
        "build_id": build_id,
        "agent": agent,
        "status": status,
        "message": message,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "artifacts": artifacts or {},
    }
    
    await redis_client.publish(channel, json.dumps(event_data))
