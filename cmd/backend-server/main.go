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

	// Start TRINI GC-aware load balancing
	fmt.Println("🔍 Starting TRINI GC-aware load balancing...")
	lb.StartTRINI()
	time.Sleep(500 * time.Millisecond)

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

// TRINI monitoring endpoints
func (h *HTTPServer) getTRINIStatus(w http.ResponseWriter, r *http.Request) {
	if h.lb.TRINI == nil {
		http.Error(w, "TRINI not initialized", http.StatusServiceUnavailable)
		return
	}

	status := map[string]interface{}{
		"active":            h.lb.TRINI.IsActive,
		"monitor_interval":  h.lb.TRINI.MonitorInterval.String(),
		"analysis_interval": h.lb.TRINI.AnalysisInterval.String(),
		"program_families":  len(h.lb.TRINI.ProgramFamilies),
		"current_policy": map[string]interface{}{
			"algorithm":         h.lb.CurrentPolicy.Algorithm,
			"gc_aware":          h.lb.CurrentPolicy.GCAware,
			"magc_threshold_ms": h.lb.CurrentPolicy.MaGCThreshold,
			"history_window":    h.lb.CurrentPolicy.HistoryWindowSize,
		},
		"servers": h.getServerTRINIDetails(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *HTTPServer) getServerTRINIDetails() []map[string]interface{} {
	servers := make([]map[string]interface{}, 0)

	for _, srv := range h.lb.Servers {
		serverInfo := map[string]interface{}{
			"server_id":          srv.ID,
			"current_family":     nil,
			"gc_history_count":   0,
			"last_magc_forecast": nil,
			"young_gen_used":     srv.YoungGenUsed,
			"old_gen_used":       srv.OldGenUsed,
			"young_gen_max":      srv.YoungGenMax,
			"old_gen_max":        srv.OldGenMax,
			"gc_count":           srv.GCCount,
			"weights":            srv.Weights,
		}

		if srv.CurrentFamily != nil {
			serverInfo["current_family"] = map[string]interface{}{
				"id":                   srv.CurrentFamily.ID,
				"name":                 srv.CurrentFamily.Name,
				"description":          srv.CurrentFamily.Description,
				"magc_threshold_ms":    srv.CurrentFamily.MaGCThreshold,
				"forecast_window_size": srv.CurrentFamily.ForecastWindowSize,
			}
		}

		serverInfo["gc_history_count"] = len(srv.GCHistory)

		if srv.LastMaGCForecast != nil {
			serverInfo["last_magc_forecast"] = map[string]interface{}{
				"predicted_time":                srv.LastMaGCForecast.PredictedTime.Format(time.RFC3339),
				"confidence":                    srv.LastMaGCForecast.Confidence,
				"young_gen_threshold":           srv.LastMaGCForecast.YoungGenThreshold,
				"time_to_magc_ms":               srv.LastMaGCForecast.TimeToMaGC,
				"forecast_created_at":           srv.LastMaGCForecast.ForecastCreatedAt.Format(time.RFC3339),
				"is_predicted_within_threshold": srv.IsMaGCPredicted(h.lb.CurrentPolicy.MaGCThreshold),
			}
		}

		servers = append(servers, serverInfo)
	}

	return servers
}

func (h *HTTPServer) getGCHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil || serverID < 1 || serverID > len(h.lb.Servers) {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	srv := h.lb.Servers[serverID-1]

	// Get query parameters for filtering
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history := srv.GCHistory
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	response := map[string]interface{}{
		"server_id":      serverID,
		"history_count":  len(srv.GCHistory),
		"returned_count": len(history),
		"gc_history":     history,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *HTTPServer) updateTRINIPolicy(w http.ResponseWriter, r *http.Request) {
	var policy server.LoadBalancingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate algorithm
	validAlgorithms := map[string]bool{"RR": true, "RAN": true, "WRR": true, "WRAN": true}
	if !validAlgorithms[policy.Algorithm] {
		http.Error(w, "Invalid algorithm. Use RR, RAN, WRR, or WRAN", http.StatusBadRequest)
		return
	}

	h.lb.SetLoadBalancingPolicy(policy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Policy updated successfully",
		"policy":  policy,
	})
}

func (h *HTTPServer) toggleTRINI(w http.ResponseWriter, r *http.Request) {
	if h.lb.TRINI == nil {
		http.Error(w, "TRINI not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Active bool `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	h.lb.TRINI.IsActive = req.Active

	status := "disabled"
	if req.Active {
		status = "enabled"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("TRINI %s successfully", status),
		"active":  h.lb.TRINI.IsActive,
	})
}

func (h *HTTPServer) getProgramFamilies(w http.ResponseWriter, r *http.Request) {
	if h.lb.TRINI == nil {
		http.Error(w, "TRINI not initialized", http.StatusServiceUnavailable)
		return
	}

	families := make(map[string]interface{})
	for id, family := range h.lb.TRINI.ProgramFamilies {
		families[id] = map[string]interface{}{
			"id":                   family.ID,
			"name":                 family.Name,
			"description":          family.Description,
			"evaluation_criteria":  family.EvaluationCriteria,
			"policy":               family.Policy,
			"forecast_window_size": family.ForecastWindowSize,
			"magc_threshold_ms":    family.MaGCThreshold,
		}
	}

	response := map[string]interface{}{
		"default_family": h.lb.TRINI.DefaultFamily.ID,
		"families":       families,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		TRINIMonitoringMiddleware(h.lb),
		GCForecastMiddleware(h.lb),
		LoadBalancingDecisionMiddleware(h.lb),
		// AuthMiddleware("your-api-key-here"), // Uncomment to enable auth
	)

	// API routes with middleware
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middlewareChain)
	api.Use(ContentTypeMiddleware) // Only for API routes

	// Original endpoints
	api.HandleFunc("/task", h.submitTask).Methods("POST")
	api.HandleFunc("/status", h.getStatus).Methods("GET")
	api.HandleFunc("/server/{id}/ping", h.pingServer).Methods("GET")

	// TRINI monitoring endpoints
	api.HandleFunc("/trini/status", h.getTRINIStatus).Methods("GET")
	api.HandleFunc("/trini/policy", h.updateTRINIPolicy).Methods("POST")
	api.HandleFunc("/trini/toggle", h.toggleTRINI).Methods("POST")
	api.HandleFunc("/trini/families", h.getProgramFamilies).Methods("GET")
	api.HandleFunc("/server/{id}/gc-history", h.getGCHistory).Methods("GET")

	// Health check (no middleware except basic ones)
	healthRouter := r.PathPrefix("/health").Subrouter()
	healthRouter.Use(Chain(RecoveryMiddleware, LoggingMiddleware))
	healthRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	fmt.Printf("🚀 HTTP Server starting on port %s\n", h.port)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  POST /api/v1/task                    - Submit a task")
	fmt.Println("  GET  /api/v1/status                  - Get system status")
	fmt.Println("  GET  /api/v1/server/{id}/ping        - Ping specific server")
	fmt.Println("  GET  /health                         - Health check")
	fmt.Println("\n🔍 TRINI GC-Aware Monitoring:")
	fmt.Println("  GET  /api/v1/trini/status            - Get TRINI status & server classifications")
	fmt.Println("  POST /api/v1/trini/policy            - Update load balancing policy")
	fmt.Println("  POST /api/v1/trini/toggle            - Enable/disable TRINI")
	fmt.Println("  GET  /api/v1/trini/families          - Get program families")
	fmt.Println("  GET  /api/v1/server/{id}/gc-history  - Get server GC history")
	fmt.Println("\n🛡️  Middleware enabled:")
	fmt.Println("  ✅ Request logging")
	fmt.Println("  ✅ CORS support")
	fmt.Println("  ✅ Rate limiting (10 req/min)")
	fmt.Println("  ✅ Panic recovery")
	fmt.Println("  ✅ Content-Type validation")
	fmt.Println("  ✅ TRINI monitoring")
	fmt.Println("  ✅ GC forecast logging")
	fmt.Println("  ✅ Load balancing decision logging")
	fmt.Println("  ⚠️  Authentication (disabled)")

	log.Fatal(http.ListenAndServe(":"+h.port, r))
}

func main() {
	port := "8080"
	server := NewHTTPServer(port)
	server.Start()
}
