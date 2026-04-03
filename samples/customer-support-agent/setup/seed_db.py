"""
Memory-efficient database seeder.
Downloads SQLite database from GCS and updates dates using SQL (no pandas).
"""

import sqlite3
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

from setup.db_config import get_db_path, is_database_seeded, DB_DIR

SQLITE_URL = "https://storage.googleapis.com/benchmarks-artifacts/travel-db/travel2.sqlite"


def download_database(dest_path: Path) -> None:
    """Download the SQLite database from Google Cloud Storage."""
    print(f"Downloading database from {SQLITE_URL}...")
    DB_DIR.mkdir(parents=True, exist_ok=True)
    urllib.request.urlretrieve(SQLITE_URL, str(dest_path))
    print("Download complete.")


def update_dates(db_path: Path) -> None:
    """
    Update dates in the database to be relative to current time.
    Uses SQL UPDATE statements directly - no pandas, no OOM.
    """
    print("Updating dates to current time...")
    conn = sqlite3.connect(str(db_path))
    cursor = conn.cursor()

    # Get the latest scheduled departure from flights
    cursor.execute("SELECT MAX(scheduled_departure) FROM flights")
    result = cursor.fetchone()
    if not result or not result[0]:
        print("No flights found, skipping date update.")
        conn.close()
        return

    latest_departure_str = result[0]

    # Parse the latest departure date
    # Handle various datetime formats
    try:
        # Try ISO format with timezone
        latest_departure = datetime.fromisoformat(
            latest_departure_str.replace("Z", "+00:00")
        )
    except ValueError:
        # Try without timezone
        latest_departure = datetime.strptime(
            latest_departure_str[:19], "%Y-%m-%d %H:%M:%S"
        ).replace(tzinfo=timezone.utc)

    # Calculate offset in days
    now = datetime.now(timezone.utc)
    offset_seconds = (now - latest_departure).total_seconds()
    offset_days = int(offset_seconds / 86400)

    print(f"Adjusting dates by {offset_days} days...")

    # Update flights table - all datetime columns
    cursor.execute(
        f"""
        UPDATE flights
        SET scheduled_departure = datetime(scheduled_departure, '{offset_days} days'),
            scheduled_arrival = datetime(scheduled_arrival, '{offset_days} days'),
            actual_departure = datetime(actual_departure, '{offset_days} days'),
            actual_arrival = datetime(actual_arrival, '{offset_days} days')
        """
    )

    # Update bookings table
    cursor.execute(
        f"""
        UPDATE bookings
        SET book_date = datetime(book_date, '{offset_days} days')
        """
    )

    conn.commit()
    conn.close()
    print("Date update complete.")


def seed_database() -> bool:
    """
    Seed the database if not already seeded.
    Returns True if seeding was performed, False if already seeded.
    """
    if is_database_seeded():
        print("Database already seeded, skipping.")
        return False

    db_path = get_db_path()

    # Download the database
    download_database(db_path)

    # Update dates to current time (memory-efficient, no pandas)
    update_dates(db_path)

    print("Database seeding complete!")
    return True


if __name__ == "__main__":
    seed_database()
