package server

import (
	"fmt"
	"math/rand"
)

// GC-Aware Round Robin (GC-RR)
func (l *LoadBalancer) GetServerGCRoundRobin(taskInput string) *Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.TRINI == nil || !l.TRINI.IsActive {
		return l.GetServerForTask(taskInput) // Fallback to regular algorithm
	}

	startIndex := l.currentServerIndex
	fTries := 0

	for fTries < len(l.Servers) {
		serverIndex := (startIndex + fTries) % len(l.Servers)
		server := l.Servers[serverIndex]

		// Check basic availability and memory capacity
		if !server.IsAvailable() || !server.CanHandleTaskSize(len(taskInput)) {
			fTries++
			continue
		}

		// GC-aware check: skip if MaGC predicted within threshold
		threshold := l.getCurrentMaGCThreshold()
		if server.IsMaGCPredicted(threshold) {
			fmt.Printf("Server %d skipped: MaGC predicted within %dms\n", server.ID, threshold)
			fTries++
			continue
		}

		// Server is suitable
		l.currentServerIndex = (serverIndex + 1) % len(l.Servers)
		fmt.Printf("Server %d selected (GC-RR)\n", server.ID)
		return server
	}

	// Escape condition: all servers have predicted MaGC, fallback to regular RR
	fmt.Println("All servers have predicted MaGC, using regular round-robin")
	return l.GetServerForTask(taskInput)
}

// GC-Aware Random (GC-RAN)
func (l *LoadBalancer) GetServerGCRandom(taskInput string) *Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.TRINI == nil || !l.TRINI.IsActive {
		return l.GetServerForTask(taskInput) // Fallback to regular algorithm
	}

	availableServers := make([]*Server, 0)

	// First, collect all available servers without predicted MaGC
	threshold := l.getCurrentMaGCThreshold()
	for _, server := range l.Servers {
		if server.IsAvailable() && server.CanHandleTaskSize(len(taskInput)) {
			if !server.IsMaGCPredicted(threshold) {
				availableServers = append(availableServers, server)
			} else {
				fmt.Printf("Server %d skipped: MaGC predicted within %dms\n", server.ID, threshold)
			}
		}
	}

	// If we have GC-safe servers, pick randomly
	if len(availableServers) > 0 {
		selectedServer := availableServers[rand.Intn(len(availableServers))]
		fmt.Printf("Server %d selected (GC-RAN)\n", selectedServer.ID)
		return selectedServer
	}

	// Escape condition: all servers have predicted MaGC, use regular random
	fmt.Println("All servers have predicted MaGC, using regular random")
	availableServers = make([]*Server, 0)
	for _, server := range l.Servers {
		if server.IsAvailable() && server.CanHandleTaskSize(len(taskInput)) {
			availableServers = append(availableServers, server)
		}
	}

	if len(availableServers) > 0 {
		return availableServers[rand.Intn(len(availableServers))]
	}

	return nil
}

// GC-Aware Weighted Round Robin (GC-WRR)
func (l *LoadBalancer) GetServerGCWeightedRoundRobin(taskInput string) *Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.TRINI == nil || !l.TRINI.IsActive {
		return l.GetServerForTask(taskInput) // Fallback to regular algorithm
	}

	// Check if all runtime weights are zero, reset if needed
	allZero := true
	for _, server := range l.Servers {
		if server.getRuntimeWeight() > 0 {
			allZero = false
			break
		}
	}

	if allZero {
		l.resetRuntimeWeights()
	}

	i := 0
	fTries := 0
	found := false
	threshold := l.getCurrentMaGCThreshold()

	for !found && fTries < len(l.Servers) {
		if i >= len(l.Servers) {
			i = 0
		}

		server := l.Servers[i]

		if server.getRuntimeWeight() > 0 {
			server.decrementRuntimeWeight()
			found = true

			// Check availability and memory
			if !server.IsAvailable() || !server.CanHandleTaskSize(len(taskInput)) {
				found = false
				server.incrementRuntimeWeight()
				i++
				fTries++
				continue
			}

			// GC-aware check
			if server.IsMaGCPredicted(threshold) {
				fmt.Printf("Server %d skipped: MaGC predicted within %dms\n", server.ID, threshold)
				found = false
				server.incrementRuntimeWeight()
				i++
				fTries++
				continue
			}

			fmt.Printf("Server %d selected (GC-WRR)\n", server.ID)
			return server
		} else {
			i++
		}
	}

	// Escape condition: fallback to regular weighted round robin
	fmt.Println("All servers have predicted MaGC, using regular weighted round-robin")
	return l.GetServerForTask(taskInput)
}

// GC-Aware Weighted Random (GC-WRAN)
func (l *LoadBalancer) GetServerGCWeightedRandom(taskInput string) *Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.TRINI == nil || !l.TRINI.IsActive {
		return l.GetServerForTask(taskInput) // Fallback to regular algorithm
	}

	threshold := l.getCurrentMaGCThreshold()

	// Calculate total weight of available servers without predicted MaGC
	totalWeight := 0
	availableServers := make([]*Server, 0)

	for _, server := range l.Servers {
		if server.IsAvailable() && server.CanHandleTaskSize(len(taskInput)) {
			if !server.IsMaGCPredicted(threshold) {
				availableServers = append(availableServers, server)
				totalWeight += server.Weights
			} else {
				fmt.Printf("Server %d skipped: MaGC predicted within %dms\n", server.ID, threshold)
			}
		}
	}

	if totalWeight == 0 || len(availableServers) == 0 {
		// Escape condition: fallback to regular weighted random
		fmt.Println("All servers have predicted MaGC, using regular weighted random")
		totalWeight = 0
		availableServers = make([]*Server, 0)
		for _, server := range l.Servers {
			if server.IsAvailable() && server.CanHandleTaskSize(len(taskInput)) {
				availableServers = append(availableServers, server)
				totalWeight += server.Weights
			}
		}

		if totalWeight == 0 || len(availableServers) == 0 {
			return nil
		}
	}

	// Weighted random selection
	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, server := range availableServers {
		currentWeight += server.Weights
		if randomWeight < currentWeight {
			fmt.Printf("Server %d selected (GC-WRAN)\n", server.ID)
			return server
		}
	}

	// Should not reach here, but return first available as fallback
	if len(availableServers) > 0 {
		return availableServers[0]
	}

	return nil
}

// GetServerGCAware is the main entry point for GC-aware load balancing
func (l *LoadBalancer) GetServerGCAware(taskInput string) *Server {
	if l.TRINI == nil || !l.TRINI.IsActive {
		return l.GetServerForTask(taskInput)
	}

	algorithm := l.CurrentPolicy.Algorithm

	switch algorithm {
	case "RR":
		return l.GetServerGCRoundRobin(taskInput)
	case "RAN":
		return l.GetServerGCRandom(taskInput)
	case "WRR":
		return l.GetServerGCWeightedRoundRobin(taskInput)
	case "WRAN":
		return l.GetServerGCWeightedRandom(taskInput)
	default:
		fmt.Printf("Unknown algorithm %s, using GC-RR\n", algorithm)
		return l.GetServerGCRoundRobin(taskInput)
	}
}

// Helper methods for weight management
func (s *Server) getRuntimeWeight() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Weights
}

func (s *Server) decrementRuntimeWeight() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Weights > 0 {
		s.Weights--
	}
}

func (s *Server) incrementRuntimeWeight() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Weights++
}

func (s *Server) resetWeight(originalWeight int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Weights = originalWeight
}

func (l *LoadBalancer) resetRuntimeWeights() {
	for _, server := range l.Servers {
		// Reset to default weight or based on server capacity
		server.resetWeight(1)
	}
}

// getCurrentMaGCThreshold returns the current MaGC threshold based on active policy
func (l *LoadBalancer) getCurrentMaGCThreshold() int64 {
	if l.TRINI == nil {
		return 2000 // Default 2 seconds
	}

	return l.CurrentPolicy.MaGCThreshold
}

// SetLoadBalancingPolicy updates the current load balancing policy
func (l *LoadBalancer) SetLoadBalancingPolicy(policy LoadBalancingPolicy) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.CurrentPolicy = policy
	fmt.Printf("Load balancing policy updated: %s (GC-aware: %t, threshold: %dms)\n",
		policy.Algorithm, policy.GCAware, policy.MaGCThreshold)
}

// AdaptPolicy adapts the load balancing policy based on current server families
func (l *LoadBalancer) AdaptPolicy() {
	if l.TRINI == nil || !l.TRINI.IsActive {
		return
	}

	// Analyze current server families and select best policy
	familyCount := make(map[string]int)
	var dominantFamily *ProgramFamily
	maxCount := 0

	for _, server := range l.Servers {
		if server.CurrentFamily != nil {
			familyCount[server.CurrentFamily.ID]++
			if familyCount[server.CurrentFamily.ID] > maxCount {
				maxCount = familyCount[server.CurrentFamily.ID]
				dominantFamily = server.CurrentFamily
			}
		}
	}

	// If we have a dominant family, use its policy
	if dominantFamily != nil && dominantFamily.Policy.GCAware {
		l.SetLoadBalancingPolicy(dominantFamily.Policy)
	}
}
