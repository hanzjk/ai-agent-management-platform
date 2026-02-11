import os
from typing import Union
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, field_validator
from agent import SimpleAgent
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

app = FastAPI(title="Simple OpenAI Agent", version="1.0.0")

# Initialize the agent
agent = SimpleAgent()


class ChatRequest(BaseModel):
    message: str
    session_id: Union[str, int]
    context: dict = None

    @field_validator('session_id')
    @classmethod
    def convert_session_id_to_string(cls, v):
        return str(v)


@app.get("/")
async def root():
    return {"message": "Simple OpenAI Agent with Traceloop is running!"}


@app.get("/health")
async def health():
    return {"status": "healthy"}


@app.post("/chat")
async def chat(request: ChatRequest):
    """
    Send a chat message
    """
    try:
        response = agent.chat(
            message=request.message,
            session_id=request.session_id,
            context=request.context
        )
        return JSONResponse(content={"response": response})
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
