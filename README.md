# Go Load Balancer System

A distributed task processing system built in Go with a React frontend, demonstrating load balancing, memory management, and real-time system monitoring.

## What is this?

This project implements a **load balancer** that distributes incoming tasks across multiple backend servers. The system intelligently routes tasks based on server availability and memory capacity, providing a practical example of distributed computing concepts.

### Demo

https://github.com/user-attachments/assets/c097b5e6-1eaa-495d-8373-108a064837e7

### Key Features

- **Round-robin load balancing** with intelligent server selection
- **Memory-aware task distribution** - servers reject tasks when memory is full
- **Garbage collection simulation** - servers automatically clean up memory
- **Real-time monitoring** - track server status, memory usage, and task distribution
- **HTTP API** with rate limiting and middleware
- **React frontend** for interactive task submission and system visualization

## How it Works

### Load Balancer Logic

1. **Task Submission**: Tasks are submitted via HTTP POST to `/api/v1/task`
2. **Server Selection**: The load balancer uses round-robin selection, but only chooses servers that:
   - Are currently available (not processing other tasks)
   - Have sufficient memory capacity for the task
3. **Task Processing**: Selected server processes the task (reverses the input string)
4. **Memory Management**: Each server tracks memory usage and triggers garbage collection at configurable thresholds
5. **Response**: Results are returned to the client with task status and output

### Architecture

```
Frontend (React) → HTTP API → Load Balancer → Server Pool (3 servers)
                      ↓
                 Middleware Chain:
                 - Rate Limiting
                 - CORS
                 - Logging
                 - Recovery
```

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 18+
- npm

### Running the Full System

1. **Start both frontend and backend**:

   ```bash
   make start
   ```

   - Backend will run on `http://localhost:8080`
   - Frontend will run on `http://localhost:5173`

2. **Or run them separately**:

   ```bash
   # Terminal 1 - Backend
   make be-start

   # Terminal 2 - Frontend
   make fe-start
   ```

### Backend Only

```bash
# Start the backend server
make be-start

# Check if it's running
curl http://localhost:8080/health
```

### Frontend Only

```bash
# Install dependencies and start
make fe-start
```

## API Endpoints

### Submit a Task

```bash
POST /api/v1/task
Content-Type: application/json

{
  "task": "hello world"
}
```

**Response**:

```json
{
  "status": "completed",
  "message": "Task processed successfully",
  "task_id": "task_123",
  "output": "dlrow olleh"
}
```

### Get System Status

```bash
GET /api/v1/status
```

**Response**:

```json
{
  "total_servers": 3,
  "available_servers": 2,
  "servers": [
    {
      "id": 1,
      "is_available": true,
      "memory_usage": 45.2,
      "tasks_processed": 12
    }
  ]
}
```

### Ping Specific Server

```bash
GET /api/v1/server/{id}/ping
```

### Health Check

```bash
GET /health
```

## Using the Frontend

The React frontend provides:

1. **Task Submission Form**: Enter tasks and submit them for processing
2. **System Status Dashboard**: Real-time view of server health and performance
3. **Task History**: See completed tasks and their results
4. **Server Monitoring**: Individual server status, memory usage, and task counts

## Configuration

### Server Configuration

Each server is configured with:

- **Memory Limit**: 100 units (configurable in `cmd/backend-server/main.go`)
- **GC Trigger**: 80% memory usage triggers garbage collection
- **Task Processing**: Simulates work by reversing input strings

### Middleware

The HTTP server includes:

- **Rate Limiting**: 10 requests per minute
- **CORS**: Cross-origin request support
- **Request Logging**: All requests are logged
- **Panic Recovery**: Graceful error handling
- **Content-Type Validation**: JSON validation for API endpoints

## Development

### Available Make Commands

**Backend**:

- `make be-start` - Start backend server
- `make be-dev` - Start with hot reload (using air)
- `make be-build` - Build binary
- `make be-stop` - Stop backend processes

**Frontend**:

- `make fe-start` - Start development server
- `make fe-build` - Build for production
- `make fe-lint` - Lint code
- `make fe-install` - Install dependencies

**General**:

- `make start` - Start both frontend and backend
- `make stop` - Stop all processes
- `make deps` - Install all dependencies
- `make clean` - Clean build artifacts
- `make help` - Show all commands

### Project Structure

```
├── cmd/backend-server/     # HTTP server implementation
│   ├── main.go            # Server setup and routes
│   └── middleware.go      # HTTP middleware
├── server/                # Load balancer core
│   ├── LoadBalancer.go    # Load balancing logic
│   ├── Server.go          # Individual server implementation
│   └── Structs.go         # Data structures
├── frontend/              # React application
│   └── src/
│       ├── components/    # UI components
│       └── store/         # State management
└── Makefile              # Build and run commands
```

## Learning Objectives

This project demonstrates:

- **Load Balancing**: How to distribute work across multiple servers
- **Memory Management**: Simulating memory pressure and garbage collection
- **HTTP APIs**: Building RESTful services with middleware
- **Real-time Monitoring**: Tracking system performance
- **Full-stack Development**: Connecting React frontend to Go backend
- **Concurrent Programming**: Handling multiple requests simultaneously

## Troubleshooting

**Backend not starting?**

- Check if port 8080 is available: `lsof -i :8080`
- Ensure Go dependencies are installed: `make be-deps`

**Frontend not connecting?**

- Verify backend is running: `curl http://localhost:8080/health`
- Check CORS configuration in middleware

**Tasks being rejected?**

- Servers may be at memory capacity - wait for garbage collection
- Check server status: `curl http://localhost:8080/api/v1/status`

## Next Steps

- Add authentication to the API
- Implement persistent task storage
- Add more sophisticated load balancing algorithms
- Create Docker containers for easy deployment
- Add metrics and monitoring dashboards
