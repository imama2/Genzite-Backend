import asyncio
import logging
import grpc
from concurrent import futures

from config import settings
from grpc_server.generated import architect_pb2_grpc
from grpc_server.architect_service import ArchitectServicer

logging.basicConfig(level=settings.LOG_LEVEL.upper())
logger = logging.getLogger(__name__)

async def serve():
    """
    Starts the gRPC server.
    """
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    
    architect_pb2_grpc.add_ArchitectServicer_to_server(
        ArchitectServicer(), server
    )
    
    listen_addr = f"[::]:{settings.GRPC_PORT}"
    server.add_insecure_port(listen_addr)
    
    logger.info(f"Starting gRPC server on {listen_addr}")
    await server.start()
    
    try:
        await server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("gRPC server is shutting down.")
        await server.stop(0)

if __name__ == "__main__":
    asyncio.run(serve())
