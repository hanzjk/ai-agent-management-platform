# Simple OpenAI Agent with Traceloop

A simple AI agent that uses OpenAI's API with Traceloop for distributed tracing.

## Features

- OpenAI GPT integration
- Traceloop SDK for observability and tracing
- FastAPI REST endpoints
- Docker support

## Prerequisites

- Python 3.11+
- OpenAI API key
- Docker (optional, for containerized deployment)

## Setup

### 1. Install Dependencies

```bash
pip install -r requirements.txt
```

### 2. Configure Environment Variables

Copy the example environment file and update it with your credentials:

```bash
cp .env.example .env
```

Edit `.env` and add your credentials:

```
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_MODEL=gpt-4o-mini
AMP_OTEL_ENDPOINT=http://localhost:4318
AMP_AGENT_API_KEY=your_amp_api_key_here
```

## Running the Agent

### Run Locally

```bash
python app.py
```

Or using uvicorn directly:

```bash
uvicorn app:app --host 0.0.0.0 --port 8000
```

### Run with Docker

Build the Docker image:

```bash
docker build -t simple-openai-agent .
```

Run the container:

```bash
docker run -p 8000:8000 \
  -e OPENAI_API_KEY=your_api_key_here \
  -e AMP_OTEL_ENDPOINT=http://host.docker.internal:4318 \
  -e AMP_AGENT_API_KEY=your_amp_api_key \
  simple-openai-agent
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8000/health
```

### Chat

```bash
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is the capital of France?",
    "session_id": "user-123"
  }'
```

With optional context:

```bash
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is the capital of France?",
    "session_id": "user-123",
    "context": {
      "user_name": "John",
      "preference": "concise answers"
    }
  }'
```

## Direct Agent Usage

You can also use the agent directly in Python:

```python
from agent import SimpleAgent

agent = SimpleAgent()

response = agent.chat(
    message="What is AI?",
    session_id="test-123",
    context={"user": "developer"}
)
print(response)
```

## Traceloop Integration

This agent uses Traceloop SDK for observability. All OpenAI API calls are automatically traced and sent to the configured endpoint.

To view traces, make sure you have an OpenTelemetry-compatible backend running (e.g., Jaeger, OpenTelemetry Collector) at the endpoint specified in `AMP_OTEL_ENDPOINT`.

## Project Structure

```
simple-openai-agent/
├── agent.py           # Core agent implementation
├── app.py             # FastAPI application
├── requirements.txt   # Python dependencies
├── Dockerfile         # Docker build configuration
├── .env.example       # Environment variables template
└── README.md          # This file
```

## License

Same as the parent project.
