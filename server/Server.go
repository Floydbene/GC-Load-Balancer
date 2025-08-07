package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func (s *Server) Start() {
	s.mu.Lock()
	s.TaskStorage = make([]string, 0)
	s.isCollectingGCTasks = false
	s.usedMemory = 0
	if s.memLimit == 0 {
		s.memLimit = 100
	}
	if s.gcPercentage == 0 {
		s.gcPercentage = 0.9 // 90%
	}
	s.mu.Unlock()
}

func (s *Server) Configure(memLimit int, gcPercentage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memLimit = memLimit
	s.gcPercentage = gcPercentage / 100.0 // Convert percentage to decimal
}

// SetMemoryLimit sets the server's memory
func (s *Server) SetMemoryLimit(limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memLimit = limit
}

// SetGCPercentage sets the GC trigger percentage (0-100)
func (s *Server) SetGCPercentage(percentage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcPercentage = percentage / 100.0 // Convert percentage to decimal
}

// GetConfiguration returns the current server configuration
func (s *Server) GetConfiguration() (memLimit int, gcPercentage float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memLimit, s.gcPercentage * 100.0 // Convert back to percentage
}

func (s *Server) CollectGCTasks() {
	s.mu.Lock()
	if s.isCollectingGCTasks {
		s.mu.Unlock()
		return
	}
	s.isCollectingGCTasks = true

	magcStartTime := time.Now()
	s.mu.Unlock()

	fmt.Printf("Server %d: Collecting GC tasks...\n", s.ID)

	gcDuration := s.calculateGCDuration()
	time.Sleep(time.Duration(gcDuration) * time.Millisecond)

	s.mu.Lock()

	magcEndTime := time.Now()
	s.MaGCDuration = magcEndTime.Sub(magcStartTime).Milliseconds()
	s.LastMaGCTime = magcEndTime
	s.GCCount++

	// Reset memory state after GC
	s.isCollectingGCTasks = false
	s.TaskStorage = make([]string, 0)
	s.usedMemory = 0
	s.YoungGenUsed = 0
	s.OldGenUsed = 0

	s.mu.Unlock()

	fmt.Printf("Server %d: GC tasks collected (duration: %dms), ready for new tasks\n",
		s.ID, s.MaGCDuration)
}

// calculateGCDuration simulates realistic GC duration based on memory usage
func (s *Server) calculateGCDuration() int64 {
	s.mu.Lock()
	memoryUsage := float64(s.usedMemory) / float64(s.memLimit)
	s.mu.Unlock()

	// Base GC duration: 10000ms to 3000ms based on memory usage
	baseDuration := 10000 + int64(memoryUsage*2500)

	// Add some randomness (±20%)
	variation := int64(float64(baseDuration) * 0.2)
	randomVariation := rand.Int63n(variation*2) - variation

	duration := baseDuration + randomVariation
	if duration < 100 {
		duration = 100
	}
	if duration > 5000 {
		duration = 5000
	}

	return duration
}

func (s *Server) IsAvailable() bool {
	s.mu.Lock()
	time.Sleep(100 * time.Millisecond)
	defer s.mu.Unlock()
	return !s.isCollectingGCTasks
}

func (s *Server) CanHandleTaskSize(taskSize int) bool {
	s.mu.Lock()
	time.Sleep(100 * time.Millisecond)

	if s.usedMemory+taskSize > s.memLimit {
		s.mu.Unlock()      // Unlock before blocking GC operation
		s.CollectGCTasks() // Remove 'go' to make it blocking
		return false
	}
	s.mu.Unlock()
	return true
}

func (s *Server) canHandleTask(input string) bool {
	s.mu.Lock()
	time.Sleep(100 * time.Millisecond)
	taskSize := len(input)
	if s.usedMemory+taskSize > s.memLimit {
		s.mu.Unlock()      // Unlock before blocking GC operation
		s.CollectGCTasks() // Remove 'go' to make it blocking
		return false
	}
	s.mu.Unlock()
	return true
}

func (s *Server) RequestTask(input string) ServiceResponse {
	// Add constant delay for server processing overhead
	time.Sleep(300 * time.Millisecond)
	resultChan := make(chan *Task, 1)

	resp := ServiceResponse{
		Status:     "pending",
		Message:    "Task received",
		TaskResult: nil,
		ResultChan: resultChan,
	}

	go func(input string) {
		if !s.IsAvailable() || !s.canHandleTask(input) {
			resultChan <- &Task{
				ID:     fmt.Sprintf("error-%d", rand.Intn(1000)),
				Input:  input,
				Output: "",
				Status: "rejected",
			}
			return
		}

		taskResult := s.handleTask(input)
		resultChan <- &taskResult

		s.mu.Lock()
		memoryUsage := float64(s.usedMemory) / float64(s.memLimit)
		gcThreshold := s.gcPercentage
		s.mu.Unlock()

		if memoryUsage >= gcThreshold {
			go s.CollectGCTasks()
		}
	}(input)

	return resp
}

func hashSHA256(s string) string {
	time.Sleep(time.Duration(rand.Intn(100)+500) * time.Millisecond)

	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

func (s *Server) handleTask(input string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskSize := len(input)
	s.usedMemory += taskSize

	// Simulate generational heap behavior
	// Most allocations go to young generation first
	youngGenAllocation := int(float64(taskSize) * 0.8) // 80% to young gen
	oldGenAllocation := taskSize - youngGenAllocation  // 20% to old gen

	s.YoungGenUsed += youngGenAllocation
	s.OldGenUsed += oldGenAllocation

	// Simulate young generation promotion to old generation
	if s.YoungGenUsed > s.YoungGenMax/2 {
		promoted := s.YoungGenUsed / 4 // Promote 25% of young gen
		s.YoungGenUsed -= promoted
		s.OldGenUsed += promoted
	}

	// Ensure we don't exceed limits
	if s.YoungGenUsed > s.YoungGenMax {
		s.YoungGenUsed = s.YoungGenMax
	}
	if s.OldGenUsed > s.OldGenMax {
		s.OldGenUsed = s.OldGenMax
	}

	task := Task{
		ID:        fmt.Sprintf("task-%d", rand.Intn(1000)),
		Input:     input,
		Output:    hashSHA256(input),
		Status:    "completed",
		CreatedAt: time.Now(),
	}

	s.TaskStorage = append(s.TaskStorage, task.ID)
	return task
}

func (s *Server) Ping() map[string]interface{} {
	s.mu.Lock()
	time.Sleep(100 * time.Millisecond)
	defer s.mu.Unlock()

	return map[string]interface{}{
		"server_id":        s.ID,
		"status":           "online",
		"is_available":     !s.isCollectingGCTasks,
		"is_collecting_gc": s.isCollectingGCTasks,
		"mem_used":         fmt.Sprintf("%.1f%%", float64(s.usedMemory)/float64(s.memLimit)*100),
		"tasks_processed":  len(s.TaskStorage),
		"task_ids":         s.TaskStorage,
		"memory_usage":     fmt.Sprintf("%d/%d (%.1f%%)", s.usedMemory, s.memLimit, float64(s.usedMemory)/float64(s.memLimit)*100),
	}
}
