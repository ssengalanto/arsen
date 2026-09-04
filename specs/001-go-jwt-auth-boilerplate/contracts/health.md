# API Contract: Health

---

## GET /healthz — Liveness Check

Returns 200 if the server process is running. No dependency checks.

**Response 200 OK**:
```json
{
  "self": "/healthz",
  "kind": "Health",
  "status": "ok"
}
```

---

## GET /readyz — Readiness Check

Returns 200 if the server is ready to serve traffic (database is reachable).

**Response 200 OK**:
```json
{
  "self": "/readyz",
  "kind": "Readiness",
  "status": "ready"
}
```

**Response 503 Service Unavailable**:
```json
{
  "self": "/readyz",
  "kind": "Readiness",
  "status": "unavailable",
  "checks": {
    "database": "unreachable"
  }
}
```
