# Real-Time Task Management System

[![Go Version](https://img.shields.io/github/go-mod/go-version/iSparshP/real-time-task-management-system)](https://go.dev/)
[![License](https://img.shields.io/github/license/iSparshP/real-time-task-management-system)](LICENSE)

## Architecture
![Image](https://github.com/user-attachments/assets/1674a9ca-c1bd-4779-bde3-4d12136293f1)

A modern, scalable real-time task management system built with Go. This system provides real-time updates, efficient task tracking, and seamless team collaboration features.

## 🌟 Features

- **Real-Time Updates**: Instant task status updates using WebSocket
- **Task Management**:
  - Create, read, update, and delete tasks
  - Task prioritization and categorization
  - Task assignment and reassignment
  - Deadline management
- **User Management**:
  - User authentication and authorization
  - Role-based access control
  - Team management
- **Project Organization**:
  - Project creation and management
  - Team workspace support
  - Project timeline tracking
- **Real-Time Notifications**:
  - Task status change alerts
  - Deadline reminders
  - Mention notifications
- **API Support**:
  - RESTful API endpoints
  - WebSocket integration
  - API documentation

## 🛠️ Technology Stack

- **Backend**: Go (99.1%)
- **Framework**: Gin
- **Database**: PostgreSQL
- **Gen Ai**: Gemini Api
- **Container**: Docker (0.9%)

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 14+

## 🚀 Getting Started

### Installation

1. Clone the repository:
```bash
git clone https://github.com/iSparshP/real-time-task-management-system.git
cd real-time-task-management-system
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env file with your configuration
```

4. Start the services using Docker Compose:
```bash
docker-compose up -d
```

5. Run the application:
```bash
go run cmd/main.go
```

### Docker Deployment

To run the entire application using Docker:

```bash
docker-compose up -d --build
```

The application will be available at `http://localhost:8080`

## 🏗️ Project Structure

```
.
├── cmd/                    # Application entry points
│   └── main.go            # Main application entry
├── internal/              # Private application and library code
│   ├── api/              # API handlers and middleware
│   ├── config/           # Configuration management
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   └── service/          # Business logic layer
├── pkg/                   # Public library code
│   ├── logger/           # Logging utilities
│   └── utils/            # Common utilities
├── migrations/           # Database migrations
├── docs/                 # Documentation
├── docker/               # Docker configurations
├── docker-compose.yml    # Docker compose configuration
├── Dockerfile           # Docker build file
└── README.md            # This file
```

## 📖 API Documentation

API documentation is available at `http://localhost:8080/swagger/index.html` when running the application.

### API Endpoints

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/tasks` - List tasks
- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks/{id}` - Get task details
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task
- `GET /api/v1/projects` - List projects
- `WS /ws/notifications` - WebSocket endpoint for real-time updates

## 🔧 Configuration

The application can be configured using environment variables or a `.env` file:

```env
# Server Configuration
SERVER_PORT=8080
ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=taskmanagement
DB_USER=postgres
DB_PASSWORD=your_password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your_jwt_secret
JWT_EXPIRATION=24h
```

## 🧪 Testing

Run the test suite:

```bash
go test ./... -v
```

Run with coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- **Sparsh Patwa** - [iSparshP](https://github.com/iSparshP)

## 🙏 Acknowledgments

- Go community for excellent libraries and tools
- Contributors who help improve this project

## 📞 Support

For support, please open an issue in the GitHub issue tracker or contact the maintainers.