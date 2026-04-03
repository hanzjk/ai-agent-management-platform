# Customer Support Agent - Deployment Guide

## Overview

The Customer Support Agent is an AI-powered customer service assistant that helps users with travel-related inquiries including flights, bookings, hotels, and car rentals. Built with LangGraph and FastAPI, this agent can search for information, make bookings, and provide comprehensive travel assistance.

**Based on**: [LangGraph Customer Support Tutorial](https://langchain-ai.github.io/langgraph/tutorials/customer-support/customer-support/#example-conversation)

## Prerequisites

Before deploying this agent, ensure you have:

### Required API Keys

- **OpenAI API Key**: For GPT-powered conversations
- **Tavily API Key**: For web search capabilities

### Database

The database is automatically set up when the agent starts:
- Downloads sample travel data from Google Cloud Storage
- Uses SQLite (no external database required)
- Idempotent: skips seeding if already initialized

## Deployment Instructions

### Step 1: Access Agent Manager

1. Navigate to the **Default** project
2. Click **"Add Agent"**
3. Select **Platform-Hosted Agent** Card

### Step 2: Configure Agent Details

Fill in the agent creation form with these exact values:

| Field                 | Value                                                   |
| --------------------- | ------------------------------------------------------- |
| **Display Name**      | `Support Agent`                                         |
| **Description**       | `AI-powered customer support agent for travel services` |
| **GitHub Repository** | `https://github.com/wso2/ai-agent-management-platform`  |
| **Branch**            | `main`                                                  |
| **App Path**          | `samples/customer-support-agent`                        |
| **Language**          | `Python`                                                |
| **Language Version**  | `3.11`                                                  |
| **Start Command**     | `python main.py`                                        |

### Step 3: Select Agent Interface

- Choose **"Chat Agent"** as the agent interface type

### Step 4: Configure Environment Variables

Add the following environment variables in the create form:

```env
OPENAI_API_KEY=<your-openai-api-key>
TAVILY_API_KEY=<your-tavily-api-key>
```

### Step 5: Deploy the Agent

1. Review all configuration details
2. Click **"Deploy"**
3. Wait for the build to complete (typically 2-5 minutes)

## Testing Your Agent

### Step 1: Navigate to Chat Interface

Click on the **"Try It"** section on the left navigation.

### Step 2: Test Sample Interactions

Try these sample questions in the chat interface:

**Flight Inquiries:**

```text
What flights do I have booked?
```

**Hotel Search:**

```text
Find me a hotel in Geneva for next week
```

### Step 3: Observe Traces

1. Click on the **"Observability"** tab on left navigation and select **Traces**
2. View traces
