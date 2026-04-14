"""
End-to-end integration tests for cherubgyre.

The server is expected to be running on http://localhost:8080 with a fresh
working directory (no users.json / followers.json / duress.json in place).

Run locally:
    JWT_SECRET=$(openssl rand -hex 32) ADMIN_TOKEN=$(openssl rand -hex 32) \
        ./cherubgyre &
    pytest test/test_api.py -v

Override the base URL or admin token via environment variables:
    BASE_URL=http://localhost:8080 ADMIN_TOKEN=... pytest test/test_api.py -v
"""

import datetime
import os
import re
import uuid

import pytest
import requests

BASE_URL = os.environ.get("BASE_URL", "http://localhost:8080")
ADMIN_TOKEN = os.environ.get("ADMIN_TOKEN", "ci-test-admin-token-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
MASTER_INVITE = "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2"

USERNAME_RE = re.compile(r"^[a-z]+-[a-z]+-[a-z]+$")


# --- Helpers ---------------------------------------------------------------


def register(normal_pin="1234", duress_pin="9876", invite=MASTER_INVITE):
    body = {"normal_pin": normal_pin, "duress_pin": duress_pin, "invite_code": invite}
    return requests.post(f"{BASE_URL}/register", json=body)


def login(username, pin):
    return requests.post(f"{BASE_URL}/login", json={"username": username, "pin": pin})


def bearer(token):
    return {"Authorization": f"Bearer {token}"}


def fresh_user(normal_pin="1234", duress_pin="9876", invite=MASTER_INVITE):
    """Register + normal login. Returns (username, token, normal_pin, duress_pin)."""
    r = register(normal_pin, duress_pin, invite)
    assert r.status_code in (200, 201), r.text
    username = r.json()["user"]["username"]
    lr = login(username, normal_pin)
    assert lr.status_code == 200, lr.text
    return username, lr.json()["token"], normal_pin, duress_pin


def duress_user(normal_pin="1234", duress_pin="9876", invite=MASTER_INVITE):
    """Register, then login with the DURESS PIN. Returns (username, duress_token)."""
    r = register(normal_pin, duress_pin, invite)
    assert r.status_code in (200, 201), r.text
    username = r.json()["user"]["username"]
    lr = login(username, duress_pin)
    assert lr.status_code == 200, lr.text
    return username, lr.json()["token"], normal_pin, duress_pin


# --- System ----------------------------------------------------------------


class TestSystem:
    def test_health(self):
        assert requests.get(f"{BASE_URL}/health").status_code == 200

    def test_root(self):
        r = requests.get(f"{BASE_URL}/")
        assert r.status_code == 200
        assert "cherubgyre" in r.text


# --- Registration & username format ---------------------------------------


class TestRegistration:
    def test_register_with_master_code(self):
        r = register()
        assert r.status_code in (200, 201), r.text
        body = r.json()
        assert "user" in body
        username = body["user"]["username"]
        assert USERNAME_RE.match(username), f"username does not match spec format: {username}"
        # PIN material must not be echoed back.
        for leak in ("normal_pin", "duress_pin", "normal_pin_hash", "duress_pin_hash"):
            assert not body["user"].get(leak), f"response leaked {leak}"

    def test_register_requires_pin_min_length(self):
        r = register(normal_pin="12", duress_pin="1234")
        assert r.status_code >= 400

    def test_register_rejects_matching_pins(self):
        r = register(normal_pin="1234", duress_pin="1234")
        assert r.status_code >= 400

    def test_register_rejects_bad_invite(self):
        r = register(invite="00000000-0000-0000-0000-000000000000")
        assert r.status_code >= 400

    def test_validate_invite_master(self):
        r = requests.post(f"{BASE_URL}/validate-invite", json={"invite_code": MASTER_INVITE})
        assert r.status_code == 200
        assert r.json().get("valid") is True


# --- Login -----------------------------------------------------------------


class TestLogin:
    def test_login_success(self):
        username, token, _, _ = fresh_user()
        assert isinstance(token, str) and token

    def test_login_wrong_pin_is_opaque(self):
        username, _, _, _ = fresh_user()
        r = login(username, "0000")
        assert r.status_code == 401

    def test_login_unknown_user_is_opaque(self):
        # Unknown-user and wrong-pin must return identical responses so an
        # attacker cannot enumerate the username space via /login.
        r1 = login("no-such-user-xyz", "1234")
        username, _, _, _ = fresh_user()
        r2 = login(username, "0000")
        assert r1.status_code == r2.status_code == 401
        assert r1.text.strip() == r2.text.strip()


# --- Profile + duress mode dummy data -------------------------------------


class TestProfileDuress:
    def test_profile_real(self):
        username, token, _, _ = fresh_user()
        r = requests.get(f"{BASE_URL}/profile", headers=bearer(token))
        assert r.status_code == 200
        body = r.json()
        assert body["username"] == username
        assert "avatar" in body

    def test_profile_unauthenticated(self):
        r = requests.get(f"{BASE_URL}/profile")
        assert r.status_code == 401

    def test_profile_duress_returns_fake_data(self):
        username, dtoken, _, _ = duress_user()
        r = requests.get(f"{BASE_URL}/profile", headers=bearer(dtoken))
        assert r.status_code == 200
        body = r.json()
        # In duress mode the real username must not be returned.
        assert body["username"] != username
        # But it should still look like a real cherubgyre username so a
        # coercer cannot spot it as fake.
        assert "avatar" in body

    def test_duress_profile_is_stable_per_user(self):
        _, dtoken, _, _ = duress_user()
        a = requests.get(f"{BASE_URL}/profile", headers=bearer(dtoken)).json()
        b = requests.get(f"{BASE_URL}/profile", headers=bearer(dtoken)).json()
        assert a == b, "duress-mode profile must be stable across requests"

    def test_duress_profile_differs_between_users(self):
        _, d1, _, _ = duress_user()
        _, d2, _, _ = duress_user()
        a = requests.get(f"{BASE_URL}/profile", headers=bearer(d1)).json()
        b = requests.get(f"{BASE_URL}/profile", headers=bearer(d2)).json()
        assert a != b, "duress-mode profile must differ between real users"

    def test_duress_followers_and_following(self):
        _, dtoken, _, _ = duress_user()
        rf = requests.get(f"{BASE_URL}/followers/ignored", headers=bearer(dtoken))
        assert rf.status_code == 200
        body = rf.json()
        assert isinstance(body, list) and len(body) >= 1

        rg = requests.get(f"{BASE_URL}/following", headers=bearer(dtoken))
        assert rg.status_code == 200
        assert isinstance(rg.json(), list)

    def test_duress_invite_is_uuid_shaped(self):
        _, dtoken, _, _ = duress_user()
        r = requests.get(f"{BASE_URL}/invite", headers=bearer(dtoken))
        assert r.status_code == 200
        code = r.json()["inviteCode"]
        uuid.UUID(code)  # raises if not a valid UUID


# --- Follow graph ----------------------------------------------------------


class TestFollowGraph:
    def test_pending_accept_unfollow(self):
        u1, t1, _, _ = fresh_user()
        u2, t2, _, _ = fresh_user()

        # u2 requests to follow u1.
        r = requests.post(f"{BASE_URL}/follow/{u1}", headers=bearer(t2))
        assert r.status_code == 200

        # u1 sees the pending request.
        r = requests.get(f"{BASE_URL}/follow/requests", headers=bearer(t1))
        assert r.status_code == 200
        assert any(entry.get("username") == u2 for entry in (r.json() or []))

        # u1 accepts.
        r = requests.post(f"{BASE_URL}/follow/accept/{u2}", headers=bearer(t1))
        assert r.status_code == 200

        # u1's followers list includes u2.
        r = requests.get(f"{BASE_URL}/followers/{u1}", headers=bearer(t1))
        assert r.status_code == 200
        assert any(entry.get("username") == u2 for entry in (r.json() or []))

        # u2 unfollows.
        r = requests.post(f"{BASE_URL}/unfollow/{u1}", headers=bearer(t2))
        assert r.status_code == 200
        r = requests.get(f"{BASE_URL}/followers/{u1}", headers=bearer(t1))
        assert not any(entry.get("username") == u2 for entry in (r.json() or []))


# --- Duress signals --------------------------------------------------------


class TestDuressSignals:
    def _post_duress(self, token, duress_pin):
        body = {
            "duress_type": "manual",
            "message": "test",
            "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
            "additional_data": {"lat": 0, "lon": 0},
            "duress_pin": duress_pin,
        }
        return requests.post(f"{BASE_URL}/duress", headers=token, json=body)

    def test_post_and_cancel(self):
        u, t, normal, duress = fresh_user()
        r = self._post_duress(bearer(t), duress)
        assert r.status_code == 200, r.text

        r = requests.get(f"{BASE_URL}/users/map", headers=bearer(t))
        assert r.status_code == 200
        assert u in (r.json() or {})

        # Cancel requires the NORMAL PIN in the body (per spec).
        r = requests.post(
            f"{BASE_URL}/duress/cancel", headers=bearer(t), json={"pin": normal}
        )
        assert r.status_code == 200, r.text

        r = requests.get(f"{BASE_URL}/users/map", headers=bearer(t))
        assert r.status_code == 200
        assert u not in (r.json() or {})

    def test_cancel_rejects_duress_pin(self):
        u, t, normal, duress = fresh_user()
        assert self._post_duress(bearer(t), duress).status_code == 200

        # Cancelling with the DURESS pin must fail — otherwise a coercer
        # holding the duress-mode session could clear the silent alert.
        r = requests.post(
            f"{BASE_URL}/duress/cancel", headers=bearer(t), json={"pin": duress}
        )
        assert r.status_code == 401

    def test_post_requires_duress_pin(self):
        _, t, _, _ = fresh_user()
        r = self._post_duress(bearer(t), "0000")
        assert r.status_code == 401

    def test_post_requires_auth(self):
        body = {
            "duress_type": "manual",
            "message": "test",
            "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
            "additional_data": {},
            "duress_pin": "9876",
        }
        assert requests.post(f"{BASE_URL}/duress", json=body).status_code == 401
        assert requests.post(f"{BASE_URL}/duress/cancel", json={"pin": "x"}).status_code == 401
        assert requests.get(f"{BASE_URL}/users/map").status_code == 401

    def test_duress_rate_limit_one_per_hour(self):
        _, t, _, duress = fresh_user()
        assert self._post_duress(bearer(t), duress).status_code == 200
        r = self._post_duress(bearer(t), duress)
        assert r.status_code == 429

    def test_silent_duress_on_duress_login(self):
        u, dtoken, _, _ = duress_user()
        # A duress-PIN login must automatically create a silent duress
        # signal visible via /users/map, even without a POST /duress.
        r = requests.get(f"{BASE_URL}/users/map", headers=bearer(dtoken))
        assert r.status_code == 200
        # The duress-mode session sees its own map via the real code path,
        # so the user's own signal should be present.
        assert u in (r.json() or {})


# --- Admin endpoint --------------------------------------------------------


class TestAdmin:
    def test_admin_requires_token(self):
        u, _, _, _ = fresh_user()
        r = requests.delete(f"{BASE_URL}/admin/users/{u}")
        assert r.status_code == 401

    def test_admin_rejects_wrong_token(self):
        u, _, _, _ = fresh_user()
        r = requests.delete(
            f"{BASE_URL}/admin/users/{u}", headers={"X-Admin-Token": "bogus"}
        )
        assert r.status_code == 401

    def test_admin_deregister_with_valid_token(self):
        u, _, _, _ = fresh_user()
        r = requests.delete(
            f"{BASE_URL}/admin/users/{u}", headers={"X-Admin-Token": ADMIN_TOKEN}
        )
        assert r.status_code == 200
        # Subsequent login must fail (user is gone).
        assert login(u, "1234").status_code == 401
