# Boxing Gym Management

A small **REST API** and **embedded web dashboard** for managing fighters, trainers, and staff users. Authentication uses **JWT**; list and mutation endpoints are protected with **role-based access** (viewer / staff / admin).

---

## Features

- **Users** — Register and login; passwords hashed with bcrypt.
- **Fighters** — CRUD-style operations with pagination; linked to a trainer.
- **Trainers** — CRUD with specializations (PostgreSQL arrays); responses include assigned fighters.
- **Web UI** — Single-page app served at `/` (English), calling `/api/v1` on the same origin.
- **Health** — `GET /healthcheck` for load balancers and monitoring.

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

| Variable     | Description                                     | Default                                                                             |
| ------------ | ----------------------------------------------- | ----------------------------------------------------------------------------------- |
| `PORT`       | HTTP listen address                             | `:3000`                                                                             |
| `DB_ADDR`    | PostgreSQL DSN (URL or `lib/pq` key=value form) | `postgres://postgres:postgres@localhost:5433/boxing-gym-management?sslmode=disable` |
| `JWT_SECRET` | HMAC secret for signing tokens                  | `secret` (change in production)                                                     |

---

## API overview

All JSON success responses use an envelope: `{ "success": true, "message": "...", "data": ... }`. Errors return `{ "error": "..." }` with an appropriate HTTP status.

Protected routes expect:

```http
Authorization: Bearer <token>
```

### Auth (no JWT required)

| Method | Path                    | Description                                                                   |
| ------ | ----------------------- | ----------------------------------------------------------------------------- |
| `POST` | `/api/v1/auth/register` | Body: `{ "email", "password" }` (min 8 chars). New users get role **viewer**. |
| `POST` | `/api/v1/auth/login`    | Body: `{ "email", "password" }`. Returns token and user (no password fields). |

### Fighters (JWT + role)

| Method   | Path                                  | Roles                | Notes                                                  |
| -------- | ------------------------------------- | -------------------- | ------------------------------------------------------ |
| `GET`    | `/api/v1/fighters/all`                | viewer, staff, admin | Query: `limit`, `page` or `offset`                     |
| `GET`    | `/api/v1/fighters/{id}`               | viewer, staff, admin |                                                        |
| `POST`   | `/api/v1/fighters/create`             | staff, admin         | JSON body: name, age, weight, wins, losses, trainer_id |
| `PUT`    | `/api/v1/fighters/profile/update?id=` | staff, admin         | Partial updates supported                              |
| `DELETE` | `/api/v1/fighters/profile/delete?id=` | staff, admin         |                                                        |

### Trainers (JWT + role)

| Method   | Path                                  | Roles                | Notes                                          |
| -------- | ------------------------------------- | -------------------- | ---------------------------------------------- |
| `GET`    | `/api/v1/trainers/all`                | viewer, staff, admin | Query: `limit`, `page` or `offset`             |
| `GET`    | `/api/v1/trainers/{id}`               | viewer, staff, admin | Includes linked fighters                       |
| `POST`   | `/api/v1/trainers/create`             | staff, admin         | Body: name, age, specialization (string array) |
| `PUT`    | `/api/v1/trainers/profile/update?id=` | staff, admin         | Partial updates                                |
| `DELETE` | `/api/v1/trainers/profile/delete?id=` | staff, admin         |                                                |

---

## Roles

| Role       | Access                                                           |
| ---------- | ---------------------------------------------------------------- |
| **viewer** | Read fighters and trainers only (default for self-registration). |
| **staff**  | Read + create / update / delete fighters and trainers.           |
| **admin**  | Same as staff (extend in code if you need admin-only features).  |

Grant **staff** or **admin** by updating the `users.role` column in the database for the relevant account.

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
  static/            # embedded dashboard (index.html)
  utils/             # JSON helpers, errors
migrations/
  init/              # runs on first Postgres container init
  003_reset_seed.sql # optional manual reset / seed (use with care)
```

---

## Migrations

- **`migrations/init/`** — Executed when the Postgres data volume is **empty** (Docker `docker-entrypoint-initdb.d`). For a clean slate, remove the Compose volume and bring the stack up again.
- **`migrations/003_reset_seed.sql`** — Optional script for resets; review before running against any database you care about.

---
