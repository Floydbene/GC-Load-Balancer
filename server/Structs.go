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
}

type LoadBalancer struct {
	mu                 sync.Mutex
	Servers            []*Server
	TaskQueue          chan string
	currentServerIndex int
}

type ServiceResponse struct {
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	TaskResult *Task      `json:"task_result,omitempty"`
	ResultChan chan *Task `json:"-"`
}
