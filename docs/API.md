# Article Service API

Base URL (local default): `http://127.0.0.1:8080`

All responses are JSON with a consistent envelope:

**Success**

```json
{
  "success": true,
  "data": {},
  "error": null
}
```

**Failure**

```json
{
  "success": false,
  "data": null,
  "error": {
    "message": "validation failed",
    "details": [
      { "field": "title", "message": "must be at least 20 characters" }
    ]
  }
}
```

`details` is always an array (possibly empty). Every response also includes an `X-Request-ID` header.

---

## Design notes (HTTP verbs)

The assessment text mentioned POST for update/delete in places. This service uses:

- **PUT** and **PATCH** for updates (not POST)
- **DELETE** for deletion (not POST)

Both choices follow common REST practice and avoid overloading POST.

---

## Endpoints

### Health probes

#### Liveness

`GET /healthz`

Returns **200** when the process is running. Does not check MySQL.

```json
{ "success": true, "data": { "status": "ok" }, "error": null }
```

#### Readiness

`GET /readyz`

Pings MySQL. Returns **200** when the database is reachable, otherwise **503**.

```json
{ "success": true, "data": { "status": "ready" }, "error": null }
```

---

### 1. Create article

`POST /article/`

**Request body**

| Field | Type | Rules |
|-------|------|--------|
| `title` | string | required, min 20 characters |
| `content` | string | required, min 200 characters |
| `category` | string | required, min 3 characters |
| `status` | string | required; one of `publish`, `draft`, `thrash` |

```json
{
  "title": "A complete guide to clean architecture",
  "content": "… at least 200 characters …",
  "category": "tech",
  "status": "draft"
}
```

**Responses**

| Status | When |
|--------|------|
| `201 Created` | Article created; `data` is the article |
| `400 Bad Request` | Validation or invalid JSON |
| `500 Internal Server Error` | Unexpected server/DB error |

---

### 2. List articles

`GET /article/{limit}/{offset}`

| Param | Rules |
|-------|--------|
| `limit` | positive integer (`> 0`) |
| `offset` | non-negative integer (`>= 0`) |

Articles are returned newest-first (`id` descending).

**Responses**

| Status | When |
|--------|------|
| `200 OK` | `data` is an array of articles (may be empty) |
| `400 Bad Request` | Invalid `limit` / `offset` |
| `500 Internal Server Error` | Unexpected server/DB error |

---

### 3. Get article by id

`GET /article/{id}`

| Param | Rules |
|-------|--------|
| `id` | positive integer |

**Responses**

| Status | When |
|--------|------|
| `200 OK` | `data` is the article |
| `400 Bad Request` | Invalid `id` |
| `404 Not Found` | No article with that id |
| `500 Internal Server Error` | Unexpected server/DB error |

---

### 4. Update article

`PUT /article/{id}`  
`PATCH /article/{id}`

Body and validation rules are identical to **Create**.

**Responses**

| Status | When |
|--------|------|
| `200 OK` | `data` is the updated article |
| `400 Bad Request` | Validation or invalid JSON / id |
| `404 Not Found` | No article with that id |
| `500 Internal Server Error` | Unexpected server/DB error |

---

### 5. Delete article

`DELETE /article/{id}`

**Responses**

| Status | When |
|--------|------|
| `200 OK` | `data` is `{ "message": "article deleted successfully" }` |
| `400 Bad Request` | Invalid `id` |
| `404 Not Found` | No article with that id |
| `500 Internal Server Error` | Unexpected server/DB error |

---

## Article resource shape

```json
{
  "id": 1,
  "title": "A complete guide to clean architecture",
  "content": "...",
  "category": "tech",
  "created_date": "2026-08-25T21:00:00+07:00",
  "updated_date": null,
  "status": "draft"
}
```

| Field | Notes |
|-------|--------|
| `created_date` | Set by MySQL `DEFAULT CURRENT_TIMESTAMP` |
| `updated_date` | `null` until an update; then set by `ON UPDATE CURRENT_TIMESTAMP` |
| `status` | `publish` \| `draft` \| `thrash` |

---

## Validation error example

`POST /article/` with an invalid body returns **400**:

```json
{
  "success": false,
  "data": null,
  "error": {
    "message": "validation failed",
    "details": [
      { "field": "title", "message": "must be at least 20 characters" },
      { "field": "content", "message": "must be at least 200 characters" },
      { "field": "category", "message": "must be at least 3 characters" },
      { "field": "status", "message": "must be one of publish, draft, thrash" }
    ]
  }
}
```

All field errors are returned together (not fail-fast on the first violation).
