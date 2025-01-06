# DiscountDB API 🚀

Backend API service for DiscountDB - the open-source coupon and discount database.

## Technology Stack 💻

- **Language**: Go
- **Cache**: Redis
- **Database**: PostgreSQL
- **API Documentation**: Swagger/OpenAPI

## Prerequisites

- Go 1.21+
- Redis
- PostgreSQL
- [swag](https://github.com/swaggo/swag) (for API documentation)

## Getting Started 🔧

1. Clone the repository

```bash
git clone https://github.com/ImGajeed76/discountdb-api.git
cd discountdb-api
```

2. Install dependencies

```bash
go mod download
```

3. Set up environment variables

```bash
cp .env.example .env
```

Configure the following in your `.env`:

- PostgreSQL connection details
- Redis connection details

4. Generate Swagger documentation

```bash
# Install swag if you haven't already
go install github.com/swaggo/swag/cmd/swag@latest
```

```bash
swag init -g ./cmd/api/main.go
```

5. Run the API

```bash
go run cmd/api/main.go
```

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