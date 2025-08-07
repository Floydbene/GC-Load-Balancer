package main

import (
	"encoding/json"
	"fmt"
	"golang_lb/server"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	lb   *server.LoadBalancer
	port string
}

type TaskRequest struct {
	Task string `json:"task"`
}

type TaskResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"`
	Output  string `json:"output,omitempty"`
}

func NewHTTPServer(port string) *HTTPServer {
	// Initialize load balancer with 3 servers
	lb := &server.LoadBalancer{
		Servers: make([]*server.Server, 0),
	}

	for i := 1; i <= 4; i++ {
		srv := &server.Server{
			ID:           i,
			LoadBalancer: lb,
			TaskStorage:  make([]string, 0),
		}
		srv.Configure(100, 80.0) // 100 memory limit, 80% GC trigger
		srv.Start()
		lb.Servers = append(lb.Servers, srv)
	}

	lb.Start()
	time.Sleep(100 * time.Millisecond)

	return &HTTPServer{
		lb:   lb,
		port: port,
	}
}

func (h *HTTPServer) submitTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Task == "" {
		http.Error(w, "Task cannot be empty", http.StatusBadRequest)
		return
	}

	srv := h.lb.GetServerForTask(req.Task)
	if srv == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TaskResponse{
			Status:  "rejected",
			Message: "No available server",
			TaskID:  fmt.Sprintf("task-%d", time.Now().UnixNano()),
		})
		return
	}

	response := srv.RequestTask(req.Task)

	// Wait for result with timeout
	select {
	case result := <-response.ResultChan:
		if result.Status == "rejected" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TaskResponse{
				Status:  "rejected",
				Message: "Server overloaded",
				TaskID:  result.ID,
			})
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TaskResponse{
				Status:  "completed",
				Message: "Task processed successfully",
				TaskID:  result.ID,
				Output:  result.Output,
			})
		}
	case <-time.After(5 * time.Second):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(TaskResponse{
			Status:  "timeout",
			Message: "Task processing timeout",
		})
	}
}

func (h *HTTPServer) getStatus(w http.ResponseWriter, r *http.Request) {
	status := make(map[string]interface{})
	servers := make([]map[string]interface{}, 0)

	availableCount := 0
	for _, srv := range h.lb.Servers {
		pingResult := srv.Ping()
		if pingResult["is_available"].(bool) {
			availableCount++
		}
		servers = append(servers, pingResult)
	}

	status["total_servers"] = len(h.lb.Servers)
	status["available_servers"] = availableCount
	status["servers"] = servers

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *HTTPServer) pingServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil || serverID < 1 || serverID > len(h.lb.Servers) {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	srv := h.lb.Servers[serverID-1]
	pingResult := srv.Ping()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingResult)
}

func (h *HTTPServer) Start() {
	r := mux.NewRouter()

	// Create rate limiter (10 requests per minute)
	rateLimiter := NewRateLimiter(10, time.Minute)

	// Apply middleware chain
	middlewareChain := Chain(
		RecoveryMiddleware,
		LoggingMiddleware,
		CORSMiddleware,
		rateLimiter.Middleware,
		// AuthMiddleware("your-api-key-here"), // Uncomment to enable auth
	)

	// API routes with middleware
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middlewareChain)
	api.Use(ContentTypeMiddleware) // Only for API routes

	api.HandleFunc("/task", h.submitTask).Methods("POST")
	api.HandleFunc("/status", h.getStatus).Methods("GET")
	api.HandleFunc("/server/{id}/ping", h.pingServer).Methods("GET")

	// Health check (no middleware except basic ones)
	healthRouter := r.PathPrefix("/health").Subrouter()
	healthRouter.Use(Chain(RecoveryMiddleware, LoggingMiddleware))
	healthRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	fmt.Printf("🚀 HTTP Server starting on port %s\n", h.port)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  POST /api/v1/task       - Submit a task")
	fmt.Println("  GET  /api/v1/status     - Get system status")
	fmt.Println("  GET  /api/v1/server/{id}/ping - Ping specific server")
	fmt.Println("  GET  /health            - Health check")
	fmt.Println("\n🛡️  Middleware enabled:")
	fmt.Println("  ✅ Request logging")
	fmt.Println("  ✅ CORS support")
	fmt.Println("  ✅ Rate limiting (10 req/min)")
	fmt.Println("  ✅ Panic recovery")
	fmt.Println("  ✅ Content-Type validation")
	fmt.Println("  ⚠️  Authentication (disabled)")

	log.Fatal(http.ListenAndServe(":"+h.port, r))
}

func main() {
	port := "8080"
	server := NewHTTPServer(port)
	server.Start()
}
