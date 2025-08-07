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
	s.mu.Unlock()

	fmt.Printf("Server %d: Collecting GC tasks...\n", s.ID)
	time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

	s.mu.Lock()
	s.isCollectingGCTasks = false
	s.TaskStorage = make([]string, 0)
	s.usedMemory = 0 // Reset memory after GC
	s.mu.Unlock()

	fmt.Printf("Server %d: GC tasks collected, ready for new tasks\n", s.ID)
}

func (s *Server) IsAvailable() bool {
	s.mu.Lock()
	time.Sleep(100 * time.Millisecond)
	defer s.mu.Unlock()
	return !s.isCollectingGCTasks
}

func (s *Server) CanHandleTaskSize(taskSize int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	time.Sleep(100 * time.Millisecond)

	if s.usedMemory+taskSize > s.memLimit {
		go s.CollectGCTasks()
		return false
	}
	return true
}

func (s *Server) canHandleTask(input string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	time.Sleep(100 * time.Millisecond)
	taskSize := len(input)
	if s.usedMemory+taskSize > s.memLimit {
		go s.CollectGCTasks()
		return false
	}
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

	s.usedMemory += len(input)

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
