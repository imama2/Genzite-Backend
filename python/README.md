# Genzite Python Agents

This directory contains the Python-based agentic AI system for the Genzite no-code website builder.

## System Components

1.  **gRPC Server (`grpc_server/`)**: Exposes the `ArchitectAgent` via a bidirectional gRPC stream for low-latency, interactive clarification with the Go orchestrator.
2.  **Build Worker (`workers/build_worker.py`)**: An asynchronous Redis consumer that orchestrates the main build pipeline (Designer → Developer → Integration → PM).
3.  **Agents (`agents/`)**: Individual, specialized agents responsible for specific tasks like design, development, and validation.
4.  **Progress Publisher (`workers/progress_publisher.py`)**: A utility for publishing real-time progress events to Redis Pub/Sub, which are then forwarded to the user's frontend.

## Setup and Installation

### Prerequisites

- Python 3.11+
- Poetry for dependency management
- Redis server running

### Installation

1.  **Navigate to the `python` directory:**
    ```bash
    cd python
    ```

2.  **Create a virtual environment (optional but recommended):**
    ```bash
    python -m venv .venv
    source .venv/bin/activate
    ```

3.  **Install dependencies using Poetry:**
    ```bash
    pip install poetry
    poetry install
    ```

4.  **Set up environment variables:**
    Copy the `.env.example` file to `.env` and fill in your details, such as API keys.
    ```bash
    cp .env.example .env
    ```

## Running the System

You need to run two main processes: the gRPC server and the build worker.

### 1. Start the gRPC Server

This server handles the interactive architect agent.

```bash
python -m grpc_server.server
```

The server will start on the port specified in your `.env` file (default: `50051`).

### 2. Start the Build Worker

This worker processes the main build jobs from the Redis queue.

```bash
python -m workers.build_worker
```

The worker will connect to Redis and wait for jobs on the `build:queue`.

### 3. Testing the Flow (Manual)

You can simulate the Go orchestrator by manually pushing a job to the Redis queue.

1.  **Connect to Redis using `redis-cli`:**
    ```bash
    redis-cli
    ```

2.  **Push a sample job to the queue:**
    Copy and paste the following command into `redis-cli`:
    ```redis
    LPUSH build:queue '{ "build_id": "test-build-123", "user_prompt": "Build me a simple portfolio website", "blueprint": { "data_models": [], "api_endpoints": [], "pages": [{"name": "Home", "path": "/", "components": ["Navbar", "Hero", "PortfolioGrid", "Footer"]}] }, "design_tokens": null, "integration_requirements": ["stripe"], "project_metadata": { "site_type": "portfolio", "language": "en" } }'
    ```

3.  **Monitor the logs:**
    You should see the build worker pick up the job and run through the agent pipeline. You can also monitor the progress events by subscribing to the Redis channel:
    ```bash
    redis-cli
    SUBSCRIBE build:test-build-123:events
    ```

## Running with Docker

The `Dockerfile` is configured to run the Python services.

1.  **Build the Docker image:**
    ```bash
    docker build -t genzite-python-agents .
    ```

2.  **Run the gRPC server:**
    ```bash
    docker run -p 50051:50051 --env-file .env genzite-python-agents python -m grpc_server.server
    ```

3.  **Run the build worker:**
    In a separate terminal:
    ```bash
    docker run --env-file .env genzite-python-agents python -m workers.build_worker
    ```
    *(Note: Ensure your Docker container can connect to your Redis instance, e.g., by using `--network="host"` or setting up a shared Docker network.)*

## Testing

The project is set up with `pytest`.

1.  **Install development dependencies:**
    ```bash
    poetry install --with dev
    ```

2.  **Run tests:**
    ```bash
    pytest
    ```
