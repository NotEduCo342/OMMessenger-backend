# OM Messenger Backend

A real-time messaging backend built with Go and Fiber framework for intranet communication.

## Features

- ✅ Real-time messaging via WebSockets
- ✅ JWT authentication
- ✅ Direct messaging
- ✅ Message delivery & read receipts
- ✅ Typing indicators
- ✅ User online status
- 🔄 Group messaging (coming soon)
- 🔄 End-to-end encryption (coming soon)
- 🔄 File sharing (coming soon)

## Tech Stack

- **Go 1.21+**
- **Fiber v2** - Web framework
- **GORM** - ORM
- **PostgreSQL** - Database
- **JWT** - Authentication
- **WebSockets** - Real-time communication

## Project Structure

```
om-backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/
│   │   ├── auth_handler.go      # Authentication endpoints
│   │   ├── message_handler.go   # Message REST endpoints
│   │   └── websocket_handler.go # WebSocket handler
│   ├── middleware/
│   │   └── auth.go              # JWT authentication middleware
│   ├── models/
│   │   ├── user.go              # User model
│   │   └── message.go           # Message model
│   ├── repository/
│   │   ├── database.go          # Database initialization
│   │   ├── user_repository.go   # User data access
│   │   └── message_repository.go# Message data access
│   └── service/
│       ├── auth_service.go      # Authentication business logic
│       └── message_service.go   # Message business logic
├── .env.example                 # Environment variables template
├── Dockerfile                   # Docker configuration
└── go.mod                       # Go module dependencies
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+
- Git

### Installation

1. **Clone the repository**
   ```bash
   cd om-backend
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and JWT secret
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Set up PostgreSQL database**
   ```bash
   createdb om_messenger
   ```

5. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

### Using Docker

1. **Build the image**
   ```bash
   docker build -t om-messenger-backend .
   ```

2. **Run with docker-compose** (recommended)
   ```bash
   docker-compose up -d
   ```

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `GET /api/users/me` - Get current user (protected)

### Messages

- `GET /api/messages?recipient_id={id}` - Get conversation (protected)
- `POST /api/messages` - Send message (protected)

### WebSocket

- `GET /ws` - WebSocket connection (protected)

## WebSocket Protocol

### Connect
```javascript
const ws = new WebSocket('ws://localhost:8080/ws', {
  headers: { Authorization: 'Bearer YOUR_JWT_TOKEN' }
});
```

### Message Types

**Send Message**
```json
{
  "type": "message",
  "recipient_id": 2,
  "content": "Hello!"
}
```

**Typing Indicator**
```json
{
  "type": "typing",
  "recipient_id": 2
}
```

**Read Receipt**
```json
{
  "type": "read",
  "message_id": 123
}
```

**Delivered Receipt**
```json
{
  "type": "delivered",
  "message_id": 123
}
```

## Development

### Run tests
```bash
go test ./...
```

### Build for production
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

## Security Considerations

- Change `JWT_SECRET` in production
- Use strong database passwords
- Enable SSL/TLS for production
- Implement rate limiting
- Consider end-to-end encryption for messages

## Contributing

This is a private project for intranet communication during internet restrictions.

## License

MIT License

## Roadmap

- [ ] Group messaging
- [ ] File uploads
- [ ] Voice messages
- [ ] End-to-end encryption
- [ ] Message search
- [ ] User blocking
- [ ] Admin panel
- [ ] Message history export

---

**Stay connected, stay safe.** 🌐
