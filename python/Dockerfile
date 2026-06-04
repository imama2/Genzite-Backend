# Use an official Python runtime as a parent image
FROM python:3.11-slim

# Set the working directory in the container
WORKDIR /app

# Copy the pyproject.toml and poetry.lock files
COPY python/pyproject.toml python/poetry.lock* /app/

# Install poetry
RUN pip install poetry

# Install dependencies
RUN poetry config virtualenvs.create false && poetry install --only main --no-root --no-interaction --no-ansi

# Copy proto files and generate Python code
COPY proto /proto
RUN mkdir -p grpc_server/generated && \
    python -m grpc_tools.protoc \
    -I/proto/agents \
    --python_out=grpc_server/generated \
    --grpc_python_out=grpc_server/generated \
    --pyi_out=grpc_server/generated \
    /proto/agents/architect.proto

# Fix the import in the generated file
RUN echo "# Generated protobuf files" > grpc_server/generated/__init__.py && \
    python -c "import sys; content = open('grpc_server/generated/architect_pb2_grpc.py').read(); \
    content = content.replace('import architect_pb2', 'from . import architect_pb2'); \
    open('grpc_server/generated/architect_pb2_grpc.py', 'w').write(content)"

# Copy the rest of the application code
COPY python/ /app/

# Expose the gRPC port
EXPOSE 50051

# Default command can be overridden
CMD ["echo", "Please specify 'grpc' or 'worker' as the command."]
