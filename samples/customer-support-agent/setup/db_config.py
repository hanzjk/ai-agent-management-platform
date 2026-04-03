"""
Database configuration module for SQLite connection.
The database is automatically seeded on first run.
"""

import sqlite3
from pathlib import Path

# Database file location (in the app directory)
DB_DIR = Path(__file__).parent.parent / "data"
DB_PATH = DB_DIR / "travel.sqlite"


def get_db_path() -> Path:
    """Get the path to the SQLite database file."""
    return DB_PATH


def get_db_connection() -> sqlite3.Connection:
    """
    Get a SQLite database connection.
    Remember to close the connection after use.
    """
    DB_DIR.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(str(DB_PATH))
    conn.row_factory = sqlite3.Row
    return conn


def is_database_seeded() -> bool:
    """Check if the database has already been seeded."""
    if not DB_PATH.exists():
        return False

    try:
        conn = get_db_connection()
        cursor = conn.cursor()
        # Check if the flights table exists and has data
        cursor.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name='flights'"
        )
        if not cursor.fetchone():
            conn.close()
            return False

        cursor.execute("SELECT COUNT(*) FROM flights")
        count = cursor.fetchone()[0]
        conn.close()
        return count > 0
    except Exception:
        return False
