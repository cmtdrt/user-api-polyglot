# UserApiPolyglot

This is a polyglot project that contains a user API written in multiple languages but using a shared database.

## Languages

![Rust](https://img.shields.io/badge/rust-%23000000.svg?style=for-the-badge&logo=rust&logoColor=white)
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

## Getting started

From the project root, run the API with Docker Compose and choose the language:

```bash
# Rust API (at http://localhost:8080)
docker compose up rust

# Go API (at http://localhost:8080)
docker compose up go
```

Run only one of the two services at a time (they use the same port). **PostgreSQL starts automatically** and is shared by both APIs.

## Database

PostgreSQL 16 runs in Docker and is common to all language implementations.

- **Port:** 5432
- **Default credentials:** user `userapi`, password `userapi`, database `userapi`
- **Connection URL:** `postgres://userapi:userapi@localhost:5432/userapi` (from the host) or `postgres://userapi:userapi@postgres:5432/userapi` (from a container)
- **Env in APIs:** `DATABASE_URL` is set automatically when running with Docker Compose.

Schema and migrations can be added in `database/init.sql`.