# Instructions for self-hosting

This guide will help you set up and run the [DiscountDB API](https://github.com/ImGajeed76/discountdb-api) with
containerized PostgreSQL and Redis services.

## Prerequisites

- [Git](https://git-scm.com/downloads)
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Go](https://golang.org/doc/install) (1.24.0 or later)

## Setup Steps

### 1. Clone the Repository

```bash
git clone https://github.com/ImGajeed76/discountdb-api.git
cd discountdb-api
```

### 2. Update Environment Variables and Compose File

Copy the `.env.example` file to `.env`:

```bash
cp .env.example .env
```

Check and update all the environment variables in the `.env` file.

If you just want a quick local setup you can use this `.env` file:

```bash
DB_USERNAME = root
DB_PASSWORD = root
DB_HOST = localhost
DB_PORT = 25060
DB_DATABASE = discountdb
DB_SSL_MODE = disable

REDIS_USERNAME = redis
REDIS_PASSWORD = root
REDIS_HOST = localhost
REDIS_PORT = 25061
REDIS_USE_TLS = false
```

Update the following in the `docker-compose.yml` file:

```yaml
POSTGRES_USER: [ DB_USERNAME ]
POSTGRES_PASSWORD: [ DB_PASSWORD ]
POSTGRES_DB: [ DB_DATABASE ]

ports:
  - "[DB_PORT]:5432"
```

```yaml
command: redis-server --requirepass [REDIS_PASSWORD]

ports:
  - "[REDIS_PORT]:6379"
```

### 3. Start the Services

Start the services using Docker Compose:

```bash
docker-compose up -d
```

This will spin up Redis and PostgreSQL containers.

### 4. Run the API Locally

Once Redis and PostgreSQL are running via Docker Compose, you can run the API:

```bash
go run cmd/api/main.go
```
