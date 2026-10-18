import os
import subprocess
import sys
from unittest.mock import MagicMock

sys.modules["kafka_client"] = MagicMock()
sys.modules["db_utils"] = MagicMock()

import bcrypt  # noqa: E402
import pytest  # noqa: E402
from fastapi.testclient import TestClient  # noqa: E402
from sqlalchemy import create_engine  # noqa: E402
from sqlalchemy.orm import sessionmaker  # noqa: E402
from sqlalchemy.pool import StaticPool  # noqa: E402

import main  # noqa: E402
from models import Base, User  # noqa: E402


@pytest.fixture
def db():
    engine = create_engine("sqlite://", poolclass=StaticPool, connect_args={"check_same_thread": False})
    Base.metadata.create_all(engine)
    session = sessionmaker(bind=engine)()
    main.app.dependency_overrides[main.get_db] = lambda: session
    yield session
    main.app.dependency_overrides.clear()
    session.close()


@pytest.fixture
def client(db):
    return TestClient(main.app)


def add_user(db, email, role, password="password1"):
    h = bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()
    db.add(User(id=email, email=email, password_hash=h, role=role))
    db.commit()


def login(client, email):
    r = client.post("/users/login", json={"email": email, "password": "password1"})
    assert r.status_code == 200
    return r


def test_login_sets_httponly_cookie(client, db):
    add_user(db, "a@x.io", "guest")
    c = login(client, "a@x.io").headers["set-cookie"].lower()
    assert "raas_token=" in c and "httponly" in c and "samesite=lax" in c


def test_admin_check_cookie_and_bearer(client, db):
    add_user(db, "adm@x.io", "admin")
    token = login(client, "adm@x.io").json()["access_token"]
    client.cookies.clear()
    client.cookies.set("raas_token", token)
    r = client.get("/users/admin-check")
    assert r.status_code == 200 and r.headers["x-auth-user"] == "adm@x.io"
    client.cookies.clear()
    assert client.get("/users/admin-check", headers={"Authorization": f"Bearer {token}"}).status_code == 200


@pytest.mark.parametrize("role", ["guest", "host"])
def test_admin_check_forbids_other_roles(client, db, role):
    add_user(db, "u@x.io", role)
    token = login(client, "u@x.io").json()["access_token"]
    assert client.get("/users/admin-check", headers={"Authorization": f"Bearer {token}"}).status_code == 403


def test_admin_check_demoted_loses_access(client, db):
    add_user(db, "adm@x.io", "admin")
    token = login(client, "adm@x.io").json()["access_token"]
    db.query(User).first().role = "guest"
    db.commit()
    assert client.get("/users/admin-check", headers={"Authorization": f"Bearer {token}"}).status_code == 403


def test_admin_check_unauthenticated_is_401(client):
    assert client.get("/users/admin-check").status_code == 401
    assert client.get("/users/admin-check", headers={"Authorization": "Bearer junk"}).status_code == 401


def test_register_admin_rejected(client):
    r = client.post("/users/register", json={"email": "e@x.io", "password": "password1", "role": "admin"})
    assert r.status_code == 422


def test_logout_clears_cookie(client):
    r = client.post("/users/logout")
    assert r.status_code == 204 and "raas_token" in r.headers["set-cookie"]


def test_seed_admin(db):
    add_user(db, "old@x.io", "admin")
    add_user(db, "taken@x.io", "host")
    main.seed_admin(db, "adm@x.io", "seedpass1")
    main.seed_admin(db, "adm@x.io", "seedpass1")  # idempotent
    roles = {u.email: u.role for u in db.query(User).all()}
    assert roles == {"old@x.io": "guest", "taken@x.io": "host", "adm@x.io": "admin"}
    # rotation: the new password takes effect for the seeded admin
    main.seed_admin(db, "adm@x.io", "rotated123")
    h = db.query(User).filter_by(email="adm@x.io").one().password_hash
    assert bcrypt.checkpw(b"rotated123", h.encode())
    # an existing non-admin with the seed email is not promoted
    main.seed_admin(db, "taken@x.io", "seedpass1")
    assert db.query(User).filter_by(email="taken@x.io").one().role == "host"


def test_seed_admin_short_password(db):
    with pytest.raises(ValueError):
        main.seed_admin(db, "adm@x.io", "short")


def test_startup_requires_jwt_secret():
    code = (
        "import sys; from unittest.mock import MagicMock; "
        "sys.modules['kafka_client']=MagicMock(); sys.modules['db_utils']=MagicMock(); import main"
    )
    env = {k: v for k, v in os.environ.items() if k != "JWT_SECRET"}
    r = subprocess.run([sys.executable, "-c", code], env=env, capture_output=True, text=True)
    assert r.returncode != 0 and "JWT_SECRET" in r.stderr
