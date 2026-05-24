import asyncio
import json
import pytest
from unittest.mock import AsyncMock, MagicMock, patch

from workers.build_worker import BuildWorker

@pytest.fixture
def mock_redis():
    """Fixture to mock the redis client."""
    mock = AsyncMock()
    # Simulate brpop returning a job and then None to stop the loop
    mock.brpop.side_effect = [
        ('build:queue', json.dumps({
            "build_id": "test-build-001",
            "user_prompt": "A test prompt",
            "blueprint": {"pages": []},
            "integration_requirements": ["stripe"]
        })),
        asyncio.CancelledError  # To stop the worker loop
    ]
    return mock

@pytest.mark.asyncio
@patch('workers.build_worker.get_redis_client')
@patch('workers.build_worker.publish_progress')
async def test_build_worker_full_pipeline(mock_publish, mock_get_redis, mock_redis):
    """
    Tests a successful run of the build worker pipeline.
    """
    mock_get_redis.return_value = mock_redis
    
    worker = BuildWorker()

    # Mock the agent run methods to return expected context updates
    worker.agents['designer'].run = AsyncMock(return_value={"design_spec": {"colors": {}}})
    worker.agents['developer'].run = AsyncMock(return_value={"code_artifacts": {"main.py": "print('hello')"}})
    worker.agents['integration'].run = AsyncMock(return_value={"integration_details": {"stripe": {"status": "ok"}}})
    worker.agents['pm'].run = AsyncMock(return_value={"validation_status": "passed"})

    try:
        await worker.start()
    except asyncio.CancelledError:
        pass

    # Assertions
    assert worker.agents['designer'].run.call_count == 1
    assert worker.agents['developer'].run.call_count == 1
    assert worker.agents['integration'].run.call_count == 1
    assert worker.agents['pm'].run.call_count == 1

    # Check that the final "COMPLETED" status was published
    final_call = mock_publish.call_args_list[-1]
    args, _ = final_call
    assert args[1] == "worker"
    assert args[2] == "COMPLETED"

@pytest.mark.asyncio
@patch('workers.build_worker.get_redis_client')
@patch('workers.build_worker.publish_progress')
async def test_build_worker_pm_failure(mock_publish, mock_get_redis, mock_redis):
    """
    Tests a pipeline run where the PM agent fails validation.
    """
    mock_get_redis.return_value = mock_redis
    mock_redis.brpop.side_effect = [
        ('build:queue', json.dumps({
            "build_id": "test-build-002",
            "user_prompt": "A test prompt",
            "blueprint": {"pages": []},
        })),
        asyncio.CancelledError
    ]

    worker = BuildWorker()
    worker.agents['designer'].run = AsyncMock(return_value={"design_spec": {}})
    worker.agents['developer'].run = AsyncMock(return_value={"code_artifacts": {}})
    worker.agents['integration'].run = AsyncMock(return_value={"integration_details": {}})
    # Simulate PM failure
    worker.agents['pm'].run = AsyncMock(return_value={"validation_status": "failed", "errors": ["Missing something"]})

    try:
        await worker.start()
    except asyncio.CancelledError:
        pass

    # PM agent should have been called
    assert worker.agents['pm'].run.call_count == 1

    # Check that the final "FAILED" status was published
    final_call = mock_publish.call_args_list[-1]
    args, _ = final_call
    assert args[1] == "worker"
    assert args[2] == "FAILED"
