import os

# main.py requires JWT_SECRET at import.
os.environ.setdefault("JWT_SECRET", "test-secret")
