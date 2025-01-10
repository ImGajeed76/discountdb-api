# DiscountDB API 🚀

Backend API service for DiscountDB - the open-source coupon and discount database.

## Technology Stack 💻

- **Language**: Go
- **Cache**: Redis
- **Database**: PostgreSQL
- **API Documentation**: Swagger/OpenAPI

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- [swag](https://github.com/swaggo/swag) (for API documentation)

## Getting Started 🛠️

### 1. Clone the repository

```bash
git clone https://github.com/ImGajeed76/discountdb-api.git
cd discountdb-api
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Set up environment variables

```bash
cp .env.example .env
```

Configure the following in your `.env`:

- PostgreSQL connection details
- Redis connection details

### 4. Generate Swagger documentation

```bash
# Install swag if you haven't already
go install github.com/swaggo/swag/cmd/swag@latest
```

```bash
swag init -g ./cmd/api/main.go
```

### 5. Run the API using Docker Compose

To simplify the setup of Redis and PostgreSQL, use the provided `docker-compose.yml` file:

1. Start the services using Docker Compose:

```bash
docker-compose up -d
```

This will spin up Redis and PostgreSQL containers.

### 6. Run the API locally

Once Redis and PostgreSQL are running via Docker Compose, you can run the API:

```bash
go run cmd/api/main.go
```

## Troubleshooting

- Ensure Docker is running and the containers are healthy (`docker ps` to check their status).
- If ports 5432 or 6379 are already in use, update the `docker-compose.yml` file to use different host ports.

## Stopping the Docker Services

To stop the services:

```bash
docker-compose down
```

This will stop and remove the containers.

## License 📜

This project is licensed under the GNU General Public License v3 (GPL-3.0). See the [LICENSE](LICENSE) file for details.

## Related Projects 🔗

- [DiscountDB Frontend](https://github.com/ImGajeed76/discountdb) - Main repository containing the SvelteKit frontend

## Contributing 🤝

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Submit a pull request

## Support 💬

For questions and support:

- Open an [issue](https://github.com/ImGajeed76/discountdb-api/issues)
- Visit the [main project repository](https://github.com/ImGajeed76/discountdb)
