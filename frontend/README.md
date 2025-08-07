# Load Balancer Frontend - Scientific Documentation

A React TypeScript web interface for a distributed task processing system implementing memory-aware round-robin load balancing with simulated garbage collection.

## System Architecture

This frontend interfaces with a **distributed computing simulation** that demonstrates load balancing algorithms, memory management, and concurrent task processing in a multi-server environment.

### Core Components

- **Load Balancer**: Implements round-robin server selection with memory capacity validation
- **Server Pool**: 4 concurrent servers with configurable memory limits and GC thresholds
- **Task Processor**: SHA-256 cryptographic hash computation with simulated processing latency
- **Memory Manager**: Garbage collection simulation with configurable trigger thresholds
- **HTTP Gateway**: RESTful API with middleware chain (rate limiting, CORS, logging, recovery)

## Load Balancing Algorithm

### Server Selection Strategy

The system implements a **memory-aware round-robin algorithm**:

```
1. Start at current server index (round-robin)
2. For each server in rotation:
   a. Check availability (not in GC mode)
   b. Validate memory capacity: usedMemory + taskSize ≤ memLimit
   c. If both conditions met, assign task and advance index
   d. If memory full, trigger asynchronous GC on that server
3. If no server available, reject task
```

### Memory Management Model

Each server maintains:

- **Memory Limit**: 100 units (configurable)
- **GC Threshold**: 80% memory usage (configurable)
- **Task Memory**: `len(input_string)` units per task
- **GC Duration**: 100-500ms random simulation

**Garbage Collection Behavior**:

- Triggered when `memoryUsage ≥ gcPercentage`
- Asynchronous execution (non-blocking)
- Complete memory reset (`usedMemory = 0`)
- Server unavailable during GC cycle

## Task Processing Pipeline

### Computational Workload

Tasks undergo **SHA-256 cryptographic hashing**:

```
Input: "hello world"
Processing: SHA-256(input) + 500-600ms latency simulation
Output: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

### Processing Latencies

- **Server Selection**: 100ms per server check
- **Task Processing**: 300ms base overhead
- **Hash Computation**: 500-600ms randomized
- **GC Collection**: 100-500ms randomized
- **Total Latency**: ~900-1500ms per task

## API Specification

### Task Submission Endpoint

```http
POST /api/v1/task
Content-Type: application/json

{
  "task": "input_string"
}
```

**Response States**:

- `completed`: Task processed successfully with SHA-256 output
- `rejected`: No server available or server overloaded
- `timeout`: Processing exceeded 5-second timeout

### System Status Monitoring

```http
GET /api/v1/status
```

**Metrics Returned**:

- Server availability states
- Memory utilization percentages
- Task processing counters
- GC collection status
- Individual server health metrics

### Server Diagnostics

```http
GET /api/v1/server/{id}/ping
```

**Diagnostic Data**:

- `is_available`: Server processing state
- `is_collecting_gc`: Garbage collection status
- `memory_usage`: Current memory utilization
- `tasks_processed`: Total task counter
- `task_ids`: Recently processed task identifiers

## Frontend Architecture

### Technology Stack

- **React 18**: Component-based UI framework
- **TypeScript**: Static type checking and development tooling
- **Vite**: Fast development server with HMR
- **Tailwind CSS**: Utility-first styling framework
- **Zustand**: Lightweight state management
- **Fetch API**: HTTP client for backend communication

### State Management

The frontend maintains:

- **Task History**: Submitted tasks with status tracking
- **System Status**: Real-time server metrics
- **Error States**: API failure handling and user feedback
- **UI State**: Form validation and loading indicators

### Real-time Monitoring

- **Auto-refresh**: 2-second polling interval for server status
- **Status Indicators**: Visual server state representation
- **Memory Visualization**: Progress bars for memory utilization
- **Task Tracking**: Status updates for submitted tasks

## Development Setup

### Prerequisites

- Node.js 18+ (ES2022 support required)
- npm 8+ (package management)
- Go 1.23+ backend server running on port 8080

### Installation & Startup

```bash
# Install dependencies
npm install

# Start development server (port 5173)
npm run dev

# Ensure backend is running
curl http://localhost:8080/health
```

### Build Configuration

```bash
# Production build
npm run build

# Build preview
npm run preview

# Code linting
npm run lint
```

## Performance Characteristics

### System Throughput

- **Rate Limit**: 10 requests/minute (configurable)
- **Concurrent Servers**: 4 parallel processing units
- **Task Capacity**: ~100 tasks per server before GC trigger
- **Processing Rate**: ~0.67-1.0 tasks/second per server

### Memory Simulation

- **Memory Model**: String length-based allocation
- **GC Efficiency**: Complete memory reclamation
- **Threshold Behavior**: Predictable GC triggering at 80% capacity
- **Recovery Time**: 100-500ms GC overhead

### Frontend Performance

- **Bundle Size**: Optimized for production deployment
- **Hot Module Replacement**: Development efficiency
- **State Updates**: Minimal re-rendering with Zustand
- **API Polling**: Efficient status monitoring

## Research Applications

This system demonstrates:

1. **Distributed Systems**: Load balancing algorithm implementation
2. **Memory Management**: Garbage collection simulation and thresholds
3. **Concurrent Programming**: Thread-safe server operations with mutexes
4. **Performance Analysis**: Latency measurement and throughput optimization
5. **System Monitoring**: Real-time metrics collection and visualization
6. **Web Architecture**: Full-stack application with RESTful API design

## Configuration Parameters

### Server Configuration

- Memory limit: 100 units (default)
- GC threshold: 80% (default)
- Server count: 4 instances
- Processing latency: 300ms base + 500-600ms hash computation

### Middleware Configuration

- Rate limiting: 10 requests/minute
- Request timeout: 5 seconds
- CORS: Enabled for cross-origin requests
- Logging: All requests logged with timestamps

---

**Note**: This is a simulation system designed for educational and research purposes. The "memory" and "garbage collection" are simplified models that demonstrate distributed systems concepts rather than actual memory management.
