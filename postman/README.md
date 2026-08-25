# Postman collection

Import `article-service.postman_collection.json` into [Postman](https://www.postman.com/) or Insomnia (Postman v2.1 format).

## Import

1. Open Postman → **Import** → select `postman/article-service.postman_collection.json`.
2. Open the collection → **Variables**.
3. Confirm `base_url` matches your running API (default `http://127.0.0.1:8080`).
   - If port 8080 is taken locally, set `APP_PORT=8081` in `.env` and use `http://127.0.0.1:8081`.

## Suggested order

1. **Create Article** — saves the new `id` into `article_id` automatically.
2. **Get Article By ID** / **List Articles**
3. **Update Article (PUT)** or **Update Article (PATCH)**
4. **Delete Article**

Optional: **Create Article (validation error example)** to inspect the 400 field-level error shape.

## Collection variables

| Variable | Default | Notes |
|----------|---------|--------|
| `base_url` | `http://127.0.0.1:8080` | API origin |
| `article_id` | `1` | Overwritten after a successful create |
| `limit` / `offset` | `10` / `0` | Pagination for list |
| `title` | ≥ 20 chars | Valid create/update title |
| `content` | ≥ 200 chars | Valid create/update content |
| `category` | `tech` | ≥ 3 chars |
| `status` | `draft` | One of `publish`, `draft`, `thrash` |
