package server

import (
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// GCSnapshot represents a point-in-time GC and memory state
type GCSnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	YoungGenUsed   int       `json:"young_gen_used"`
	OldGenUsed     int       `json:"old_gen_used"`
	YoungGenMax    int       `json:"young_gen_max"`
	OldGenMax      int       `json:"old_gen_max"`
	TotalMemUsed   int       `json:"total_mem_used"`
	TotalMemMax    int       `json:"total_mem_max"`
	GCCount        int       `json:"gc_count"`
	LastMaGCTime   time.Time `json:"last_magc_time"`
	MaGCDuration   int64     `json:"magc_duration_ms"`
	IsCollectingGC bool      `json:"is_collecting_gc"`
}

// MaGCForecast represents a predicted Major GC event
type MaGCForecast struct {
	PredictedTime     time.Time `json:"predicted_time"`
	Confidence        float64   `json:"confidence"`
	YoungGenThreshold int       `json:"young_gen_threshold"`
	TimeToMaGC        int64     `json:"time_to_magc_ms"`
	ForecastCreatedAt time.Time `json:"forecast_created_at"`
}

// ProgramFamily defines GC characteristics and policies
type ProgramFamily struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	EvaluationCriteria map[string]interface{} `json:"evaluation_criteria"`
	Policy             LoadBalancingPolicy    `json:"policy"`
	ForecastWindowSize int                    `json:"forecast_window_size"`
	MaGCThreshold      int64                  `json:"magc_threshold_ms"`
}

// LoadBalancingPolicy defines the rules for load balancing
type LoadBalancingPolicy struct {
	Algorithm         string `json:"algorithm"` // RR, RAN, WRR, WRAN
	GCAware           bool   `json:"gc_aware"`
	MaGCThreshold     int64  `json:"magc_threshold_ms"`
	HistoryWindowSize int    `json:"history_window_size"`
}

// TRINI represents the TRINI adaptive system
type TRINI struct {
	mu               sync.RWMutex
	ProgramFamilies  map[string]*ProgramFamily `json:"program_families"`
	DefaultFamily    *ProgramFamily            `json:"default_family"`
	MonitorInterval  time.Duration             `json:"monitor_interval"`
	AnalysisInterval time.Duration             `json:"analysis_interval"`
	IsActive         bool                      `json:"is_active"`
}

type Server struct {
	mu                  sync.Mutex
	TaskQueue           chan Task
	ID                  int
	LoadBalancer        *LoadBalancer
	TaskStorage         []string
	isCollectingGCTasks bool
	usedMemory          int
	memLimit            int
	gcPercentage        float64 // GC trigger percentage (0.0-1.0)

	// TRINI GC-aware extensions
	GCHistory        []GCSnapshot   `json:"gc_history"`
	CurrentFamily    *ProgramFamily `json:"current_family"`
	LastMaGCForecast *MaGCForecast  `json:"last_magc_forecast"`
	YoungGenUsed     int            `json:"young_gen_used"`
	OldGenUsed       int            `json:"old_gen_used"`
	YoungGenMax      int            `json:"young_gen_max"`
	OldGenMax        int            `json:"old_gen_max"`
	GCCount          int            `json:"gc_count"`
	LastMaGCTime     time.Time      `json:"last_magc_time"`
	MaGCDuration     int64          `json:"magc_duration_ms"`
	Weights          int            `json:"weights"` // For weighted algorithms
}

type LoadBalancer struct {
	mu                 sync.Mutex
	Servers            []*Server
	TaskQueue          chan string
	currentServerIndex int

	// TRINI extensions
	TRINI         *TRINI              `json:"trini"`
	CurrentPolicy LoadBalancingPolicy `json:"current_policy"`
}

type ServiceResponse struct {
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	TaskResult *Task      `json:"task_result,omitempty"`
	ResultChan chan *Task `json:"-"`
}
