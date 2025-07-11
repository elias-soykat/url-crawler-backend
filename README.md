# URL Crawler Backend

A high-performance Go backend service for crawling and analyzing web pages. The service provides RESTful APIs for URL management, authentication, and detailed web page analysis including link validation, HTML structure analysis, and form detection.

## Features

- **Web Page Crawling**: Crawl and analyze web pages for various metrics
- **Link Analysis**: Detect internal/external links and validate broken links
- **HTML Analysis**: Extract title, HTML version, and heading structure
- **Form Detection**: Identify login forms on web pages
- **Authentication**: JWT-based authentication system
- **RESTful API**: Clean and well-documented API endpoints
- **Database Storage**: MySQL-based data persistence with proper indexing
- **Background Processing**: Asynchronous URL processing with worker pools
- **Docker Support**: Complete containerization with Docker and Docker Compose

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│   Gin Router    │───▶│   API Handlers  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Middleware    │    │  Crawler Service│
                       │   (Auth/CORS)   │    │  (Worker Pool)  │
                       └─────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Database      │    │   HTTP Client   │
                       │   (MySQL/GORM)  │    │   (goquery)     │
                       └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)

### Using Docker Compose

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-crawler-backend
   ```

2. **Create environment file**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the services**

   ```bash
   docker-compose up -d
   ```

4. **Seed the database with admin user**

   ```bash
   docker-compose --profile seed up seed
   ```

5. **Access the services**
   - API: http://localhost:8080
   - Adminer (Database UI): http://localhost:8081

### Local Development

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Set up MySQL database**

   ```bash
   # Using Docker
   docker run -d --name mysql \
     -e MYSQL_ROOT_PASSWORD=root_password \
     -e MYSQL_DATABASE=url_crawler \
     -e MYSQL_USER=crawler_user \
     -e MYSQL_PASSWORD=crawler_password \
     -p 3306:3306 mysql:8.0
   ```

3. **Set environment variables**

   ```bash
   export MYSQL_HOST=localhost
   export MYSQL_PORT=3306
   export MYSQL_DATABASE=url_crawler
   export MYSQL_USER=crawler_user
   export MYSQL_PASSWORD=crawler_password
   export JWT_SECRET=your-secret-key
   ```

4. **Run the application**

   ```bash
   go run main.go
   ```

5. **Seed the database**
   ```bash
   go run cmd/seed/main.go
   ```

## API Documentation

### Authentication

#### POST /auth/login

Authenticate a user and receive a JWT token.

**Request:**

```json
{
  "username": "admin",
  "password": "adminpass"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T12:00:00Z",
  "user_id": 1,
  "username": "admin"
}
```

### URL Management

#### POST /urls

Add a new URL for crawling.

**Headers:** `Authorization: Bearer <token>`

**Request:**

```json
{
  "address": "https://example.com"
}
```

**Response:**

```json
{
  "id": 1,
  "address": "https://example.com",
  "status": "queued",
  "created_at": "2024-01-01T12:00:00Z"
}
```

#### GET /urls

List URLs with pagination and search.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

- `page`: Page number (default: 1)
- `size`: Page size (default: 10, max: 100)
- `sort`: Sort field (default: "created_at desc")
- `q`: Search query
- `status`: Filter by status

**Response:**

```json
{
  "data": [...],
  "page": 1,
  "size": 10,
  "total": 100,
  "pages": 10
}
```

#### GET /urls/:id

Get detailed information about a specific URL.

**Headers:** `Authorization: Bearer <token>`

**Response:**

```json
{
  "id": 1,
  "address": "https://example.com",
  "title": "Example Domain",
  "html_version": "HTML5",
  "heading_counts": { "h1": 1, "h2": 2 },
  "internal_links": 5,
  "external_links": 3,
  "broken_links": 1,
  "broken_list": [{ "url": "https://broken.com", "code": "404" }],
  "has_login_form": false,
  "status": "done",
  "created_at": "2024-01-01T12:00:00Z"
}
```

#### POST /urls/bulk

Perform bulk operations on URLs.

**Headers:** `Authorization: Bearer <token>`

**Request:**

```json
{
  "action": "rerun",
  "ids": [1, 2, 3]
}
```

**Actions:**

- `rerun`: Reset URLs to queued status for re-crawling
- `delete`: Delete URLs

### Health Check

#### GET /health

Check service health status.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "url-crawler"
}
```

## Configuration

### Environment Variables

| Variable         | Default            | Description        |
| ---------------- | ------------------ | ------------------ |
| `MYSQL_HOST`     | `localhost`        | MySQL host         |
| `MYSQL_PORT`     | `3306`             | MySQL port         |
| `MYSQL_DATABASE` | `url_crawler`      | Database name      |
| `MYSQL_USER`     | `crawler_user`     | Database user      |
| `MYSQL_PASSWORD` | `crawler_password` | Database password  |
| `JWT_SECRET`     | `changeme`         | JWT signing secret |
| `JWT_DURATION`   | `24h`              | JWT token duration |
| `PORT`           | `8080`             | HTTP server port   |

### Database Schema

#### Users Table

- `id`: Primary key
- `username`: Unique username
- `password`: Hashed password
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

#### URLs Table

- `id`: Primary key
- `address`: URL address (unique)
- `title`: Page title
- `html_version`: HTML version detected
- `heading_counts`: JSON of heading counts
- `internal_links`: Number of internal links
- `external_links`: Number of external links
- `broken_links`: Number of broken links
- `broken_list`: JSON array of broken links
- `has_login_form`: Boolean flag
- `status`: Processing status (queued/running/done/error)
- `error`: Error message if failed
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

## Development

### Project Structure

```
go-crawler-backend/
├── cmd/
│   └── seed/           # Database seeding utility
├── internal/
│   ├── api/            # HTTP handlers
│   ├── crawler/        # Web crawling logic
│   ├── db/             # Database models and operations
│   └── middleware/     # HTTP middleware
├── scripts/            # Database initialization scripts
├── main.go            # Application entry point
├── Dockerfile         # Container definition
├── docker-compose.yml # Service orchestration
└── README.md          # This file
```

### Code Quality Guidelines

1. **Error Handling**: Always handle errors explicitly
2. **Logging**: Use structured logging for debugging
3. **Validation**: Validate all inputs
4. **Security**: Use proper authentication and authorization
5. **Performance**: Use connection pooling and proper indexing
6. **Testing**: Write unit and integration tests

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Development build
go build -o url-crawler main.go

# Production build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o url-crawler main.go
```

## Deployment

### Docker

```bash
# Build image
docker build -t url-crawler .

# Run container
docker run -d \
  -p 8080:8080 \
  -e MYSQL_HOST=your-mysql-host \
  -e MYSQL_PASSWORD=your-password \
  -e JWT_SECRET=your-secret \
  url-crawler
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Monitoring

### Health Checks

The service provides health check endpoints:

- `GET /health`: Basic health status
- Docker health checks are configured for all services

### Logging

The application logs important events:

- Server startup/shutdown
- Database operations
- Crawling activities
- Authentication attempts
- Error conditions

### Metrics

Consider adding metrics collection for:

- Request rates and response times
- Database query performance
- Crawling success/failure rates
- Queue depths and processing times

## Security Considerations

1. **JWT Secrets**: Use strong, unique secrets in production
2. **Database Security**: Use dedicated database users with minimal privileges
3. **Network Security**: Restrict database access to application servers
4. **Input Validation**: All inputs are validated and sanitized
5. **HTTPS**: Use HTTPS in production environments
6. **Rate Limiting**: Consider implementing rate limiting for API endpoints

## Troubleshooting

### Common Issues

1. **Database Connection Failed**

   - Check MySQL service is running
   - Verify connection credentials
   - Ensure network connectivity

2. **Crawling Failures**

   - Check target URL accessibility
   - Verify network connectivity
   - Review error logs for specific issues

3. **Authentication Issues**
   - Verify JWT secret is consistent
   - Check token expiration
   - Ensure proper Authorization header format

### Logs

```bash
# View application logs
docker-compose logs backend

# View database logs
docker-compose logs mysql

# Follow logs in real-time
docker-compose logs -f backend
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
