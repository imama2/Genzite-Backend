# Use an official Python runtime as a parent image
FROM python:3.11-slim

# Set the working directory in the container
WORKDIR /app

# Copy the pyproject.toml and poetry.lock files
COPY pyproject.toml poetry.lock* /app/

# Install poetry
RUN pip install poetry

# Install dependencies
RUN poetry config virtualenvs.create false && poetry install --no-dev --no-interaction --no-ansi

# Copy the rest of the application code
COPY . /app

# Expose the gRPC port
EXPOSE 50051

# Default command can be overridden
CMD ["echo", "Please specify 'grpc' or 'worker' as the command."]
