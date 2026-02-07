# TaskMate - Simple Task Management API

TaskMate is a lightweight task management application with a REST API and web interface. Manage your tasks with due dates, priorities, and status tracking - all stored in a simple JSON file.

![TaskMate UI](img/homepage.png)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Git (optional)

### Installation & Running

1. **Clone or download the repository:**
   ```bash
   git clone <repository-url>
   cd taskmate
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Create configuration file:**
   ```bash
   cp config.json.example config.json
   ```
   
   **Important:** The example config includes a default password. Change it immediately for security (see [Security](#changing-the-master-password) section).

4. **Start the server:**
   ```bash
   go run main.go
   ```

5. **Open your browser:**
   
   Visit **http://localhost:8080** to access the web interface.

That's it! You're ready to start managing tasks.

## Using TaskMate

### Web Interface

The web UI provides an intuitive interface to:
- ➕ Create tasks with title, description, due date, and priority
- 📋 View all tasks or filter by status (pending/completed)
- ✓ Mark tasks as complete or reopen them
- ✏️ Edit task details
- 🗑️ Delete tasks

### API Usage

TaskMate provides a REST API for programmatic access.

#### Step 1: Generate an API Token

To create, update, or delete tasks via API, you need a token. Generate one using the master password:

```bash
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"password": "your_password_here"}'
```

Response:
```json
{
  "token": "a1b2c3d4e5f6...",
  "message": "Token generated successfully. Save this token securely, it won't be shown again."
}
```

**Important:** Save this token! It's only shown once.

**Note:** The default password is set in `config.json`. For production use, change it immediately (see [Security](#changing-the-master-password) section).

#### Step 2: Use the API

**View tasks (no authentication needed):**
```bash
# Get all tasks
curl http://localhost:8080/api/v1/tasks

# Get pending tasks only
curl http://localhost:8080/api/v1/tasks/pending

# Get specific task
curl http://localhost:8080/api/v1/tasks/1
```

**Create a task (requires token):**
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "X-API-Token: YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Deploy to Production",
    "description": "Deploy v2.0 release",
    "due_date": "2024-12-31",
    "priority": "high"
  }'
```

**Update a task (requires token):**
```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "X-API-Token: YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Deploy to Production",
    "description": "Deploy v2.0 release",
    "due_date": "2024-12-31",
    "priority": "high",
    "status": "completed"
  }'
```

**Delete a task (requires token):**
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1 \
  -H "X-API-Token: YOUR_TOKEN_HERE"
```

## API Reference

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | None |
| POST | `/api/v1/auth/token` | Generate API token | Password |
| GET | `/api/v1/tasks` | Get all tasks | None |
| GET | `/api/v1/tasks/pending` | Get pending tasks only | None |
| GET | `/api/v1/tasks/{id}` | Get specific task | None |
| POST | `/api/v1/tasks` | Create new task | Token |
| PUT | `/api/v1/tasks/{id}` | Update task | Token |
| DELETE | `/api/v1/tasks/{id}` | Delete task | Token |

## Security

### Authentication

TaskMate uses a two-tier authentication system:

1. **Password Authentication** - Required only when generating tokens
2. **Token Authentication** - Required for creating, updating, or deleting tasks

**Note:** Reading tasks (GET requests) doesn't require authentication.

**Security Warning:** The example configuration includes a default password hash. Always change this before deploying to production or exposing the application to the internet.

### Changing the Master Password

**Important:** Change the default password before deploying to production!

1. Generate a SHA-256 hash of your new password:
   ```bash
   # On macOS/Linux
   echo -n "your_new_password" | shasum -a 256
   
   # On Windows (PowerShell)
   $password = "your_new_password"
   $hash = [System.Security.Cryptography.SHA256]::Create().ComputeHash([System.Text.Encoding]::UTF8.GetBytes($password))
   [System.BitConverter]::ToString($hash).Replace("-", "").ToLower()
   ```

2. Update the `password_hash` field in `config.json` with the generated hash.

3. Restart the server.

### Security Features

- Passwords and tokens are never stored in plain text
- SHA-256 hashing for all sensitive data
- Multiple tokens supported for different users/applications
- Thread-safe operations

## Docker Deployment

### Using Docker

```bash
# Build the image
docker build -t taskmate:latest .

# Run the container
docker run -d \
  -p 8080:8080 \
  -e TASKMATE_PASSWORD_HASH=ea424017c57b0d0b2f262edd821dca2dc3cfcbb47e296a9007415af86bbc6ac1 \
  -v $(pwd)/tasks.json:/app/tasks.json \
  --name taskmate \
  taskmate:latest
```

### Using Docker Compose

```bash
docker-compose up -d
```

### Environment Variables

For containerized deployments, you can use environment variables:

- `TASKMATE_PORT` - Server port (default: 8080)
- `TASKMATE_PASSWORD_HASH` - SHA-256 hash of master password

Generate a password hash:
```bash
echo -n "your_secure_password" | shasum -a 256
```

## Configuration

TaskMate uses `config.json` for configuration:

```json
{
  "api_key": "add-token",
  "port": "8080",
  "password_hash": "ea424017c57b0d0b2f262edd821dca2dc3cfcbb47e296a9007415af86bbc6ac1",
  "token_hashes": []
}
```

- `port` - Server port
- `password_hash` - SHA-256 hash of master password
- `token_hashes` - Array of generated token hashes (managed automatically)

**Configuration Priority:**
1. Environment variables (highest)
2. config.json file
3. Default values (lowest)

## Data Storage

Tasks are stored in `tasks.json` in the current directory. The file is automatically created and updated as you manage tasks.

Example:
```json
[
  {
    "id": 1,
    "title": "Deploy to Production",
    "description": "Deploy v2.0 release",
    "due_date": "2024-12-31",
    "priority": "high",
    "status": "pending",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

## Development

### Running Tests

```bash
go test -v ./...
```

### Building from Source

```bash
go build -o taskmate
./taskmate
```

## What You'll Learn (For Developers)

This project demonstrates:
- Building REST APIs in Go
- HTTP routing with Gorilla Mux
- JSON serialization and file persistence
- Token-based authentication
- Thread-safe data structures
- Concurrent request handling
- Docker containerization

## Architecture Overview (For Developers)

### Core Components

1. **Task Struct** - Represents a task with metadata
2. **TaskStore** - Thread-safe storage with JSON persistence
3. **Server** - HTTP server with authentication middleware
4. **Router** - URL routing and endpoint handling

### Key Concepts

**Thread-Safe Storage:**
- Uses `sync.RWMutex` for concurrent access
- Allows multiple readers or one writer
- Prevents data corruption

**Authentication Middleware:**
- Token validation for write operations
- SHA-256 hash comparison
- No authentication for read operations

**JSON Persistence:**
- Automatic save after modifications
- Load on startup
- Human-readable format

## License

MIT License - see LICENSE file for details.