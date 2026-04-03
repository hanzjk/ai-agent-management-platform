from setup.seed_db import seed_database

# Seed the database on startup (idempotent - skips if already seeded)
seed_database()

from app import app

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
