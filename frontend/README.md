# Load Balancer Frontend

A React TypeScript application for interacting with the Go load balancer backend.

## Features

- 🚀 Submit tasks to the load balancer
- 📊 Real-time system status monitoring
- 🖥️ Individual server status tracking
- 📝 Task history with status indicators
- 🎨 Modern UI with Tailwind CSS
- 🔄 Auto-refreshing server status

## Tech Stack

- **React 18** with TypeScript
- **Vite** for fast development
- **Tailwind CSS** for styling
- **Zustand** for state management
- **Fetch API** for backend communication

## Getting Started

1. **Install dependencies:**

   ```bash
   npm install
   ```

2. **Start the development server:**

   ```bash
   npm run dev
   ```

3. **Make sure the Go backend is running on port 8080:**

   ```bash
   # In the root directory
   go run cmd/backend-server/*.go
   ```

4. **Open your browser:**
   ```
   http://localhost:5173
   ```

## Usage

1. **Submit Tasks:** Enter any text in the task input field and click "Submit Task"
2. **Monitor Status:** View real-time server status in the right panel
3. **View Results:** See completed tasks with their outputs in the task list
4. **Error Handling:** Any errors will be displayed at the top of the page

## API Integration

The frontend communicates with the Go backend through these endpoints:

- `POST /api/v1/task` - Submit a new task
- `GET /api/v1/status` - Get system status
- `GET /api/v1/server/{id}/ping` - Ping specific server

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
