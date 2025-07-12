# URL Crawler Backend

A high-performance Go backend service for crawling and analyzing web pages with JWT authentication, MySQL database, and Docker support.

## Features

- **Web Page Crawling** - Analyze web pages for various metrics (REAL-TIME FEATURE IS NOT DEVELOP YET DUE TO TIME CONSTRAINTS)
- **Link Analysis** - Detect internal/external links and validate broken links
- **HTML Analysis** - Extract title, HTML version, and heading structure
- **Authentication** - JWT-based user authentication
- **Docker Ready** - Complete containerization with Docker Compose
- **RESTful API** - Clean and documented API endpoints

## Stack

- Go
- Gin
- MySQL
- Docker
- JWT

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) (Programming Language)
- [Docker](https://docs.docker.com/get-docker/) (Containerization)
- [MySQL](https://dev.mysql.com/downloads/) (Database)
- [Make](https://www.gnu.org/software/make/) (Build and Automation)

### 1. Clone & Setup

```bash
# Clone the repository
git clone git@github.com:elias-soykat/url-crawler-backend.git
cd url-crawler-backend

# Create environment file
cp .env.example .env
```

### 2. Available Commands

```bash
# Show all available commands 
make help

# Development
make docker-build    # Build Docker image
make docker-run      # Start services
make docker-stop     # Stop services
make restart         # Restart services

# Monitoring
make logs           # View all logs
make logs-backend   # View backend logs
make logs-mysql     # View MySQL logs
make status         # Check service status

# Maintenance
make docker-clean   # Clean Docker resources
```

### 3. API Endpoints

#### Register User

```bash
POST /auth/signup
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

#### Login

```bash
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

#### URL Management

> **Note**: All URL endpoints require JWT token in Authorization header: `Bearer <token>`

#### Submit Single URL

```bash
POST /urls
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://example.com"
}
```

#### List URLs

```bash
GET /urls?page=1&limit=10
Authorization: Bearer <token>
```

#### Get URL Details

```bash
GET /urls/:id
Authorization: Bearer <token>
```

### 4. Development Workflow

#### Making Changes

1. Make your code changes
2. Rebuild and restart:
   ```bash
   make docker-build
   make restart
   ```

#### Database Access

Access database via [Adminer](https://www.adminer.org/) web interface:

- URL: http://localhost:8081
- System: MySQL
- Server: mysql
- Username: `MYSQL_USER` from .env
- Password: `MYSQL_PASSWORD` from .env
- Database: `MYSQL_DATABASE` from .env

### 5. Architecture

```
HTTP Client → Gin Router → API Handlers
                ↓              ↓
           Middleware → Crawler Service
                ↓              ↓
           Database ← HTTP Client
```
