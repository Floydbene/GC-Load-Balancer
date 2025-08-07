package server

import "fmt"

func (l *LoadBalancer) Start() {
	l.TaskQueue = make(chan string)

	go func() {
		for task := range l.TaskQueue {
			server := l.GetServerForTask(task)
			if server != nil {
				server.RequestTask(task)
			} else {
				fmt.Printf("❌ No server can handle task: '%s'\n", task)
			}
		}
	}()
	for i := range l.Servers {
		go l.Servers[i].Start()
	}
}

// Legacy method for backward compatibility (when task size unknown)
func (l *LoadBalancer) GetServer() *Server {
	return l.GetServerForTask("")
}

// New method that considers both availability and memory capacity
func (l *LoadBalancer) GetServerForTask(taskInput string) *Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	startIndex := l.currentServerIndex
	for i := 0; i < len(l.Servers); i++ {
		serverIndex := (startIndex + i) % len(l.Servers)
		server := l.Servers[serverIndex]

		// Check both availability and memory capacity
		if server.IsAvailable() && server.CanHandleTaskSize(len(taskInput)) {
			fmt.Printf("Server %d is available and can handle task (round-robin)\n", server.ID)
			l.currentServerIndex = (serverIndex + 1) % len(l.Servers)
			return server
		} else if !server.IsAvailable() {
			fmt.Printf("Server %d is busy/unavailable\n", server.ID)
		} else {
			fmt.Printf("Server %d is available but memory full\n", server.ID)
		}
	}

	fmt.Println("No server can handle this task")
	return nil
}
