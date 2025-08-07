package main

import (
	"golang_lb/server"
	"log"
	"net/http"
	"sync"
	"time"
)

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Skip logging for status endpoint
		if r.URL.Path != "/api/v1/status" {
			duration := time.Since(start)
			log.Printf("%s %s %v", r.Method, r.URL.Path, duration)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware implements simple rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr

		rl.mu.Lock()
		now := time.Now()

		// Clean old requests
		if requests, exists := rl.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[clientIP] = validRequests
		}

		// Check rate limit
		if len(rl.requests[clientIP]) >= rl.limit {
			rl.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Rate limit exceeded"}`))
			return
		}

		// Add current request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware provides simple API key authentication
func AuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "API key required"}`))
				return
			}

			if providedKey != apiKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid API key"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware ensures JSON content type for API endpoints
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Content-Type must be application/json"}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// TRINIMonitoringMiddleware logs TRINI-specific events and decisions
func TRINIMonitoringMiddleware(lb *server.LoadBalancer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code and response
			wrapped := &triniResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				lb:             lb,
				requestStart:   start,
				requestPath:    r.URL.Path,
				requestMethod:  r.Method,
			}

			// Log TRINI state before request
			if r.URL.Path == "/api/v1/task" && r.Method == "POST" {
				logTRINIPreRequest(lb)
			}

			next.ServeHTTP(wrapped, r)

			// Log TRINI state after request for task submissions
			if r.URL.Path == "/api/v1/task" && r.Method == "POST" {
				duration := time.Since(start)
				logTRINIPostRequest(lb, wrapped.statusCode, duration)
			}
		})
	}
}

// triniResponseWriter extends responseWriter to capture TRINI-specific data
type triniResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	lb            *server.LoadBalancer
	requestStart  time.Time
	requestPath   string
	requestMethod string
}

func (trw *triniResponseWriter) WriteHeader(code int) {
	trw.statusCode = code
	trw.ResponseWriter.WriteHeader(code)
}

func logTRINIPreRequest(lb *server.LoadBalancer) {
	if lb.TRINI == nil || !lb.TRINI.IsActive {
		log.Printf("🔍 TRINI: Inactive - using regular load balancing")
		return
	}

	availableServers := 0
	gcPredictedServers := 0

	for _, srv := range lb.Servers {
		if srv.IsAvailable() {
			availableServers++
			if srv.IsMaGCPredicted(lb.CurrentPolicy.MaGCThreshold) {
				gcPredictedServers++
			}
		}
	}

	log.Printf("🔍 TRINI: Policy=%s, Available=%d, GC-Predicted=%d, Threshold=%dms",
		lb.CurrentPolicy.Algorithm, availableServers, gcPredictedServers, lb.CurrentPolicy.MaGCThreshold)
}

func logTRINIPostRequest(lb *server.LoadBalancer, statusCode int, duration time.Duration) {
	if lb.TRINI == nil || !lb.TRINI.IsActive {
		return
	}

	// Log server family classifications
	familyCounts := make(map[string]int)
	for _, srv := range lb.Servers {
		if srv.CurrentFamily != nil {
			familyCounts[srv.CurrentFamily.Name]++
		} else {
			familyCounts["Unclassified"]++
		}
	}

	log.Printf("🔍 TRINI: Request completed in %v (status: %d)", duration, statusCode)
	log.Printf("🔍 TRINI: Family distribution: %v", familyCounts)
}

// GCForecastMiddleware logs detailed GC forecasting information
func GCForecastMiddleware(lb *server.LoadBalancer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log GC forecasts for task submissions
			if r.URL.Path == "/api/v1/task" && r.Method == "POST" && lb.TRINI != nil && lb.TRINI.IsActive {
				logGCForecasts(lb)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func logGCForecasts(lb *server.LoadBalancer) {
	for _, srv := range lb.Servers {
		if srv.LastMaGCForecast != nil {
			forecast := srv.LastMaGCForecast
			timeUntilMaGC := time.Until(forecast.PredictedTime)

			if timeUntilMaGC > 0 && timeUntilMaGC.Milliseconds() <= lb.CurrentPolicy.MaGCThreshold {
				log.Printf("🔮 Server %d: MaGC predicted in %v (confidence: %.2f)",
					srv.ID, timeUntilMaGC, forecast.Confidence)
			}
		}
	}
}

// LoadBalancingDecisionMiddleware logs which server was selected and why
func LoadBalancingDecisionMiddleware(lb *server.LoadBalancer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/task" && r.Method == "POST" {
				// This will be logged by the load balancing algorithms themselves
				// but we can add additional context here
				if lb.TRINI != nil && lb.TRINI.IsActive {
					log.Printf("⚖️  Using GC-aware %s algorithm", lb.CurrentPolicy.Algorithm)
				} else {
					log.Printf("⚖️  Using regular round-robin algorithm")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Chain combines multiple middleware functions
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
