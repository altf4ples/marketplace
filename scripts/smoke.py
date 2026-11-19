#!/usr/bin/env python3
"""End-to-end smoke test. Run after `docker compose up --build`."""
import json
import os
import urllib.error
import urllib.request
import uuid
from decimal import Decimal

BASE = os.getenv("API_URL", "http://localhost:8080")

def request(method, path, payload=None):
    data = None if payload is None else json.dumps(payload).encode()
    req = urllib.request.Request(
        BASE + path,
        data=data,
        method=method,
        headers={"Content-Type": "application/json"} if data is not None else {},
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as exc:
        body = exc.read().decode(errors="replace")
        raise RuntimeError(f"{method} {path}: HTTP {exc.code}: {body}") from exc

def check(condition, message):
    if not condition:
        raise RuntimeError(message)

def main():
    request("GET", "/healthz")
    suffix = uuid.uuid4().hex[:8]
    _, user = request("POST", "/users", {"name": "Smoke Tester"})
    user_id = user["id"]
    _, store = request("POST", "/store", {"name": "Smoke " + suffix})
    store_id = store["id"]
    _, product = request(
        "POST", f"/stores/{store_id}/products",
        {"name": "Smoke Product", "description": "Created by smoke test", "price": "12.34"},
    )
    product_id = product["id"]

    _, cart = request("POST", "/cart/items", {
        "user_id": user_id, "product_id": product_id, "quantity": 2,
    })
    check(len(cart["items"]) == 1 and cart["items"][0]["quantity"] == 2, "cart add failed")

    _, order = request("POST", "/orders", {"user_id": user_id})
    order_id = order["id"]
    check(Decimal(str(order["total_price"])) == Decimal("24.68"), "incorrect order total")
    check(len(order["items"]) == 1, "order item snapshot missing")
    check(order["items"][0]["product_name"] == "Smoke Product", "product name snapshot missing")
    check(Decimal(str(order["items"][0]["price"])) == Decimal("12.34"), "price snapshot missing")

    _, cart = request("GET", f"/cart?user_id={user_id}")
    check(cart["items"] == [], "cart was not emptied after checkout")

    for event, expected in [
        ("assemble", "assembling"),
        ("ship", "shipped"),
        ("deliver", "delivered"),
    ]:
        _, order = request(
            "POST", f"/orders/{order_id}/events",
            {"user_id": user_id, "event": event},
        )
        check(order["status"] == expected, f"unexpected status after {event}: {order['status']}")

    _, detail = request("GET", f"/orders/{order_id}?user_id={user_id}")
    check(detail["status"] == "delivered", "final order state mismatch")
    check(len(detail["history"]) == 4, "order history should contain created + 3 transitions")

    _, notifications = request("GET", f"/users/{user_id}/notifications")
    check(len(notifications["notifications"]) >= 4, "expected checkout and transition notifications")
    print("Smoke test passed: user -> store -> product -> cart -> checkout -> delivered -> notifications")

if __name__ == "__main__":
    main()
