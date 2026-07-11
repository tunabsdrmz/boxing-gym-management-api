# 🥊 Boxing Gym Management API

A production-ready REST API designed for managing boxing gyms, members, trainers, attendance, and daily gym operations.

Built with **Go**, **Chi Router**, and **PostgreSQL**, the project follows a clean layered architecture with JWT authentication, role-based authorization, OpenAPI documentation, rate limiting, logging, and Docker support.

---

## ✨ Features

- 🔐 JWT Authentication & Refresh Tokens
- 👥 Role-Based Access Control (Admin, Staff, Viewer)
- 🥊 Member & Trainer Management
- 📅 Attendance Tracking
- 📊 Dashboard Statistics
- 📄 CSV Export
- 📖 OpenAPI / Swagger Documentation
- ⚡ Rate Limiting
- 📝 Request Logging
- ❤️ Health Check Endpoint
- 🐳 Docker Compose Support
- 🗄 PostgreSQL Database
- 📄 Environment-Based Configuration

---

## 🛠 Tech Stack

| Category         | Technologies          |
| ---------------- | --------------------- |
| Language         | Go                    |
| Router           | Chi                   |
| Database         | PostgreSQL            |
| Authentication   | JWT                   |
| Authorization    | RBAC                  |
| Documentation    | OpenAPI / Swagger     |
| Containerization | Docker Compose        |
| Configuration    | Environment Variables |

---

## 🏗 Architecture

```text
                Client
                   │
                   ▼
            Chi HTTP Router
                   │
        ┌──────────┴──────────┐
        │                     │
   Middleware           Route Handlers
        │                     │
        └──────────┬──────────┘
                   ▼
              Service Layer
                   ▼
           Repository Layer
                   ▼
              PostgreSQL
```

---

## 📁 Project Structure

```
cmd/
internal/
    handlers/
    middleware/
    services/
    repositories/
    models/
pkg/
docs/
scripts/
```

---

## 🚀 Getting Started

### Clone

```bash
git clone https://github.com/yourusername/boxing-gym-management-api.git
```

### Install

```bash
go mod download
```

### Configure

Create a `.env` file using `.env.example`.

### Start PostgreSQL

```bash
docker compose up -d
```

### Run

```bash
go run ./cmd/server
```

---

## 📚 API Documentation

After starting the application, OpenAPI documentation is available through the configured Swagger endpoint.

---

## 🔒 Authentication

The API uses JWT Authentication.

Available roles include:

- Admin
- Staff
- Viewer

Protected endpoints require a valid Bearer Token.

---

## 📸 Screenshots

> Screenshots will be added soon.

Suggested screenshots:

- Dashboard
- Swagger UI
- Login
- Member Management
- Trainer Management
- Attendance
- CSV Export

---

## 📈 Future Improvements

- Unit & Integration Tests
- CI/CD Pipeline
- Prometheus Metrics
- Grafana Monitoring
- Kubernetes Deployment
- Audit Logs
- Email Notifications

---

## 📄 License

This project was built for educational and portfolio purposes.
