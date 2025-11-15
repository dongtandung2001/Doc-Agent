from concurrent import futures

import grpc
import ai_service_pb2_grpc

from ai_service_impl import AIServiceServicer
from config import Config
from logger import setup_logger

logger = setup_logger(__name__)


def serve():
    """Start the gRPC server."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    # Add servicer to server using generated gRPC code
    ai_service_pb2_grpc.add_AIServiceServicer_to_server(
        AIServiceServicer(), server
    )

    port = Config.GRPC_PORT
    server.add_insecure_port(f'[::]:{port}')

    server.start()
    logger.info(f"AI Service started on port {port}")

    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Shutting down AI Service")
        server.stop(0)


if __name__ == '__main__':
    serve()
