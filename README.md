# Boxing Gym Management

A small **REST API** and **embedded web dashboard** for managing fighters, trainers, and staff users. Authentication uses **JWT**; list and mutation endpoints are protected with **role-based access** (viewer / staff / admin).

---

## Features

- **Users** — Register and login; bcrypt; short-lived **access JWT** + **refresh token**; optional password reset flow; **locked** accounts.
- **Fighters** — CRUD, pagination, filters, CSV export, extended profile (KVKK-sensitive fields optional), **assistant/corner** trainers.
- **Trainers** — CRUD with specializations (PostgreSQL arrays); responses include assigned fighters.
- **Operations** — Schedule (overlap checks), daily **attendance**, **announcements**.
- **Admin API** — List users, change role, lock/unlock (`admin` only).
- **Web UI** — Dashboard at `/`; auto **refresh** on `401` when a refresh token exists.
- **Docs** — `GET /api/v1/openapi.yaml`.
- **Health** — `GET /healthcheck` for load balancers and monitoring.
- **Hardening** — Optional **CORS** and per-IP **rate limit**; chi **request ID** + request logging.

---

## Tech stack

| Layer       | Choice                               |
| ----------- | ------------------------------------ |
| Language    | Go 1.26+                             |
| HTTP router | [chi](https://github.com/go-chi/chi) |
| Database    | PostgreSQL (via `lib/pq`)            |
| Auth        | JWT (HS256), custom RBAC middleware  |

---

## Prerequisites

- [Go](https://go.dev/dl/) **1.26.1** or newer (see `go.mod`).
- [Docker](https://docs.docker.com/get-docker/) (optional, for PostgreSQL via Compose).

---

## Quick start

### 1. Database

Start PostgreSQL with the bundled Compose file. On first boot, scripts under `migrations/init/` run automatically and create schema, seed data, and users.

```bash
docker compose up -d
```

Default DB URL matches the app default: `postgres://postgres:postgres@localhost:5433/boxing-gym-management?sslmode=disable`.

### 2. Configuration

Create a `.env` file in the project root (see [Environment variables](#environment-variables)). At minimum you need a reachable `DB_ADDR` and a strong `JWT_SECRET` in production.

Example:

```env
PORT=:3000
DB_ADDR=postgres://postgres:postgres@localhost:5433/boxing-gym-management?sslmode=disable
JWT_SECRET=your-secret-here
```

`.env` is listed in `.gitignore`. If it was ever committed, stop tracking it with `git rm --cached .env`.

### 3. Run the server

```bash
go run ./cmd
```

The process loads `.env` via `godotenv`. If `.env` is missing, startup will log a fatal error — create the file first.

- **Dashboard:** [http://localhost:3000/](http://localhost:3000/) (or your `PORT`)
- **API base:** `http://localhost:3000/api/v1`

### Live reload (optional)

If you use [Air](https://github.com/air-verse/air), configuration is in `.air.toml`:

```bash
air
```

---

## Environment variables

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PORT` | HTTP listen address | `:3000` |
| `DB_ADDR` | PostgreSQL DSN (URL or `lib/pq` key=value form) | `postgres://postgres:postgres@localhost:5433/boxing-gym-management?sslmode=disable` |
| `JWT_SECRET` | HMAC secret for signing access tokens | `secret` (change in production) |
| `JWT_ACCESS_TTL_MINUTES` | Access JWT lifetime | `15` |
| `JWT_REFRESH_TTL_DAYS` | Refresh token storage lifetime | `7` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated allowed browser origins (empty = CORS middleware off) | *(empty)* |
| `RATE_LIMIT_RPM` | Max requests per minute per IP (`0` = off) | `0` |
| `DEV_RETURN_PASSWORD_RESET_TOKEN` | If `true`, `POST /auth/forgot-password` includes `reset_token` in JSON (dev only) | `false` |

---

## API overview

All JSON success responses use an envelope: `{ "success": true, "message": "...", "data": ... }`. Errors return `{ "error": "..." }` with an appropriate HTTP status.

Protected routes expect:

```http
Authorization: Bearer <access_token>
```

OpenAPI sketch: `GET /api/v1/openapi.yaml` (import into Swagger UI or an editor).

### Auth (no JWT required)

| Method | Path | Description |
| ------ | ---- | ----------- |
| `POST` | `/api/v1/auth/register` | `{ "email", "password" }` (min 8). Role **viewer**. Returns `access_token`, `refresh_token`, `token` (alias of access), `expires_in`, `user`. |
| `POST` | `/api/v1/auth/login` | Same response shape as register. Locked users get `403`. |
| `POST` | `/api/v1/auth/refresh` | `{ "refresh_token" }` — rotates refresh, returns new pair. |
| `POST` | `/api/v1/auth/forgot-password` | `{ "email" }` — always `200`; creates reset token (see `DEV_RETURN_PASSWORD_RESET_TOKEN`). |
| `POST` | `/api/v1/auth/reset-password` | `{ "token", "new_password" }` — revokes refresh sessions for that user. |

### Fighters (JWT + role)

| Method | Path | Roles | Notes |
| ------ | ---- | ----- | ----- |
| `GET` | `/api/v1/fighters/all` | viewer+ | `limit`, `page` / `offset`, `q`, `weight_class`, `fighter_status`, `sort` (`name` \| `weight` \| `created_at`), `order` (`asc` \| `desc`). Includes `assistant_trainers`. |
| `GET` | `/api/v1/fighters/export` | viewer+ | CSV; same filters as `/all`, cap 10k rows. |
| `GET` | `/api/v1/fighters/{id}` | viewer+ | Profile + assistants. |
| `POST` | `/api/v1/fighters/{id}/assistants` | staff, admin | `{ "trainer_id", "role": "assistant" \| "corner" }` |
| `DELETE` | `/api/v1/fighters/{id}/assistants/{trainerID}` | staff, admin | |
| `POST` | `/api/v1/fighters/create` | staff, admin | Extended optional: `health_notes`, `contract_end`, emergency contacts, `weight_class`, `fighter_status`, `license_number`. |
| `PUT` | `/api/v1/fighters/profile/update?id=` | staff, admin | Partial updates |
| `DELETE` | `/api/v1/fighters/profile/delete?id=` | staff, admin | |

### Trainers (JWT + role)

| Method   | Path                                  | Roles                | Notes                                          |
| -------- | ------------------------------------- | -------------------- | ---------------------------------------------- |
| `GET`    | `/api/v1/trainers/all`                | viewer, staff, admin | Query: `limit`, `page` or `offset`             |
| `GET`    | `/api/v1/trainers/{id}`               | viewer, staff, admin | Includes linked fighters                       |
| `POST`   | `/api/v1/trainers/create`             | staff, admin         | Body: name, age, specialization (string array) |
| `PUT`    | `/api/v1/trainers/profile/update?id=` | staff, admin         | Partial updates                                |
| `DELETE` | `/api/v1/trainers/profile/delete?id=` | staff, admin         |                                                |

### Operations (JWT)

| Method | Path | Roles | Notes |
| ------ | ---- | ----- | ----- |
| `GET` | `/api/v1/schedule/events` | viewer+ | Query `from`, `to` (RFC3339) |
| `GET` | `/api/v1/schedule/events/{id}` | viewer+ | |
| `POST` | `/api/v1/schedule/events` | staff, admin | Ring/mat/general bookings; overlap rejected |
| `PUT` | `/api/v1/schedule/events/{id}` | staff, admin | |
| `DELETE` | `/api/v1/schedule/events/{id}` | staff, admin | |
| `GET` | `/api/v1/attendance` | viewer+ | `?gym_date=YYYY-MM-DD` (default today UTC) |
| `POST` | `/api/v1/attendance` | staff, admin | Upsert attendance row |
| `DELETE` | `/api/v1/attendance/{id}` | staff, admin | |
| `GET` | `/api/v1/announcements/active` | viewer+ | Non-expired |
| `GET` | `/api/v1/announcements/all` | staff, admin | |
| `POST` | `/api/v1/announcements` | staff, admin | |
| `PUT` | `/api/v1/announcements/{id}` | staff, admin | `expires_at` empty string clears expiry |
| `DELETE` | `/api/v1/announcements/{id}` | staff, admin | |

### Admin (JWT, **admin** only)

| Method | Path | Description |
| ------ | ---- | ----------- |
| `GET` | `/api/v1/admin/users` | Paginated user list (`limit`, `page` / `offset`) |
| `PATCH` | `/api/v1/admin/users/{id}` | `{ "role"?, "locked"?, "locked_reason"? }` |

---

## Roles

| Role       | Access |
| ---------- | ------ |
| **viewer** | Read fighters, trainers, schedule, attendance, announcements (active). |
| **staff**  | Viewer + mutations on fighters, trainers, schedule, attendance, announcements. |
| **admin**  | Staff + `GET/PATCH /admin/users`. |

You can still change roles in SQL; the admin API updates `users` without a direct DB client.

---

## Project layout

```
cmd/                 # application entrypoint and HTTP mounting
internal/
  auth/              # JWT, claims, context helpers
  config/            # global app config
  db/                # database connection
  env/               # env helpers
  handler/           # HTTP handlers
  middleware/        # auth + RBAC
  repository/        # SQL/data access
  routes/            # chi route registration
  apidocs/           # embedded OpenAPI YAML
  static/            # embedded dashboard (index.html)
  utils/             # JSON helpers, errors
migrations/
  init/              # runs on first Postgres container init (includes `005_features.sql` on new DBs)
  003_reset_seed.sql # optional manual reset / seed (use with care)
```

---

## Migrations

- **`migrations/init/`** — Executed when the Postgres data volume is **empty** (Docker `docker-entrypoint-initdb.d`). For a clean slate, remove the Compose volume and bring the stack up again.
- **`005_features.sql`** — Adds operations tables, fighter profile columns, refresh/password-reset tokens, `users.locked`. On an **existing** database that was created before this file existed, run that SQL manually once (or migrate with your own tool).
- **`migrations/003_reset_seed.sql`** — Optional script for resets; review before running against any database you care about.

---
