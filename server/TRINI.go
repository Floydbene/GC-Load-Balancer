package server

import (
	"fmt"
	"math"
	"time"
)

// NewTRINI creates a new TRINI adaptive system
func NewTRINI() *TRINI {
	trini := &TRINI{
		ProgramFamilies:  make(map[string]*ProgramFamily),
		MonitorInterval:  2 * time.Second,
		AnalysisInterval: 10 * time.Second,
		IsActive:         true,
	}

	// Initialize default program families
	trini.initializeDefaultFamilies()

	return trini
}

// initializeDefaultFamilies sets up predefined program families
func (t *TRINI) initializeDefaultFamilies() {
	// Short MaGC duration family (< 500ms)
	shortMaGCFamily := &ProgramFamily{
		ID:          "short-magc",
		Name:        "Short MaGC Duration",
		Description: "Applications with MaGC events typically under 500ms",
		EvaluationCriteria: map[string]interface{}{
			"max_magc_duration": 500,
			"min_samples":       5,
		},
		Policy: LoadBalancingPolicy{
			Algorithm:         "RR",
			GCAware:           true,
			MaGCThreshold:     1000, // 1 second
			HistoryWindowSize: 20,
		},
		ForecastWindowSize: 15,
		MaGCThreshold:      1000,
	}

	// Medium MaGC duration family (500ms - 2s)
	mediumMaGCFamily := &ProgramFamily{
		ID:          "medium-magc",
		Name:        "Medium MaGC Duration",
		Description: "Applications with MaGC events between 500ms and 2s",
		EvaluationCriteria: map[string]interface{}{
			"max_magc_duration": 2000,
			"min_magc_duration": 500,
			"min_samples":       5,
		},
		Policy: LoadBalancingPolicy{
			Algorithm:         "WRR",
			GCAware:           true,
			MaGCThreshold:     3000, // 3 seconds
			HistoryWindowSize: 30,
		},
		ForecastWindowSize: 25,
		MaGCThreshold:      3000,
	}

	// Long MaGC duration family (> 2s)
	longMaGCFamily := &ProgramFamily{
		ID:          "long-magc",
		Name:        "Long MaGC Duration",
		Description: "Applications with MaGC events over 2 seconds",
		EvaluationCriteria: map[string]interface{}{
			"min_magc_duration": 2000,
			"min_samples":       3,
		},
		Policy: LoadBalancingPolicy{
			Algorithm:         "WRR",
			GCAware:           true,
			MaGCThreshold:     5000, // 5 seconds
			HistoryWindowSize: 40,
		},
		ForecastWindowSize: 35,
		MaGCThreshold:      5000,
	}

	// Default family for new/unclassified applications
	defaultFamily := &ProgramFamily{
		ID:          "default",
		Name:        "Default",
		Description: "Default family for unclassified applications",
		EvaluationCriteria: map[string]interface{}{
			"min_samples": 0,
		},
		Policy: LoadBalancingPolicy{
			Algorithm:         "RR",
			GCAware:           false,
			MaGCThreshold:     2000,
			HistoryWindowSize: 10,
		},
		ForecastWindowSize: 10,
		MaGCThreshold:      2000,
	}

	t.ProgramFamilies["short-magc"] = shortMaGCFamily
	t.ProgramFamilies["medium-magc"] = mediumMaGCFamily
	t.ProgramFamilies["long-magc"] = longMaGCFamily
	t.ProgramFamilies["default"] = defaultFamily
	t.DefaultFamily = defaultFamily
}

// StartTRINI starts the TRINI monitoring and analysis loops
func (lb *LoadBalancer) StartTRINI() {
	if lb.TRINI == nil {
		lb.TRINI = NewTRINI()
	}

	// Initialize servers with default family
	for _, server := range lb.Servers {
		server.initializeTRINI(lb.TRINI.DefaultFamily)
	}

	// Start monitoring loop
	go lb.monitoringLoop()

	// Start analysis loop
	go lb.analysisLoop()

	fmt.Println("🔍 TRINI GC-aware load balancing started")
}

// monitoringLoop periodically collects GC data from servers
func (lb *LoadBalancer) monitoringLoop() {
	ticker := time.NewTicker(lb.TRINI.MonitorInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !lb.TRINI.IsActive {
			continue
		}

		for _, server := range lb.Servers {
			go server.collectGCSnapshot()
		}
	}
}

// analysisLoop periodically analyzes GC patterns and updates program families
func (lb *LoadBalancer) analysisLoop() {
	ticker := time.NewTicker(lb.TRINI.AnalysisInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !lb.TRINI.IsActive {
			continue
		}

		for _, server := range lb.Servers {
			go server.analyzeAndAdapt(lb.TRINI)
		}
	}
}

// initializeTRINI initializes a server with TRINI capabilities
func (s *Server) initializeTRINI(defaultFamily *ProgramFamily) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GCHistory = make([]GCSnapshot, 0, 100)
	s.CurrentFamily = defaultFamily
	s.YoungGenMax = s.memLimit / 2 // Assume 50% for young generation
	s.OldGenMax = s.memLimit / 2   // Assume 50% for old generation
	s.Weights = 1                  // Default weight for weighted algorithms
}

// collectGCSnapshot captures current GC and memory state
func (s *Server) collectGCSnapshot() {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := GCSnapshot{
		Timestamp:      time.Now(),
		YoungGenUsed:   s.YoungGenUsed,
		OldGenUsed:     s.OldGenUsed,
		YoungGenMax:    s.YoungGenMax,
		OldGenMax:      s.OldGenMax,
		TotalMemUsed:   s.usedMemory,
		TotalMemMax:    s.memLimit,
		GCCount:        s.GCCount,
		LastMaGCTime:   s.LastMaGCTime,
		MaGCDuration:   s.MaGCDuration,
		IsCollectingGC: s.isCollectingGCTasks,
	}

	// Add to history (keep last 100 snapshots)
	s.GCHistory = append(s.GCHistory, snapshot)
	if len(s.GCHistory) > 100 {
		s.GCHistory = s.GCHistory[1:]
	}
}

// analyzeAndAdapt analyzes GC patterns and adapts program family if needed
func (s *Server) analyzeAndAdapt(trini *TRINI) {
	s.mu.Lock()
	gcHistory := make([]GCSnapshot, len(s.GCHistory))
	copy(gcHistory, s.GCHistory)
	currentFamily := s.CurrentFamily
	s.mu.Unlock()

	if len(gcHistory) < 3 {
		return // Need minimum samples for analysis
	}

	// Evaluate current family suitability
	if !s.evaluateCurrentFamily(gcHistory, currentFamily) {
		// Find better family
		newFamily := s.findBestFamily(gcHistory, trini)
		if newFamily != nil && newFamily.ID != currentFamily.ID {
			s.mu.Lock()
			s.CurrentFamily = newFamily
			s.mu.Unlock()
			fmt.Printf("Server %d: Adapted to program family '%s'\n", s.ID, newFamily.Name)
		}
	}

	// Generate MaGC forecast
	forecast := s.generateMaGCForecast(gcHistory)
	if forecast != nil {
		s.mu.Lock()
		s.LastMaGCForecast = forecast
		s.mu.Unlock()
	}
}

// evaluateCurrentFamily checks if current family still suits the server
func (s *Server) evaluateCurrentFamily(history []GCSnapshot, family *ProgramFamily) bool {
	if family == nil {
		return false
	}

	criteria := family.EvaluationCriteria
	minSamples, _ := criteria["min_samples"].(int)

	if len(history) < minSamples {
		return len(history) == 0 // Only valid if no samples yet
	}

	// Calculate recent MaGC durations
	recentDurations := make([]int64, 0)
	for i := len(history) - 1; i >= 0 && len(recentDurations) < 10; i-- {
		if history[i].MaGCDuration > 0 {
			recentDurations = append(recentDurations, history[i].MaGCDuration)
		}
	}

	if len(recentDurations) == 0 {
		return true // No MaGC events yet
	}

	avgDuration := int64(0)
	for _, d := range recentDurations {
		avgDuration += d
	}
	avgDuration /= int64(len(recentDurations))

	// Check against family criteria
	if maxDuration, exists := criteria["max_magc_duration"].(int); exists {
		if avgDuration > int64(maxDuration) {
			return false
		}
	}

	if minDuration, exists := criteria["min_magc_duration"].(int); exists {
		if avgDuration < int64(minDuration) {
			return false
		}
	}

	return true
}

// findBestFamily finds the most suitable program family for the server
func (s *Server) findBestFamily(history []GCSnapshot, trini *TRINI) *ProgramFamily {
	trini.mu.RLock()
	defer trini.mu.RUnlock()

	// Calculate recent MaGC durations
	recentDurations := make([]int64, 0)
	for i := len(history) - 1; i >= 0 && len(recentDurations) < 10; i-- {
		if history[i].MaGCDuration > 0 {
			recentDurations = append(recentDurations, history[i].MaGCDuration)
		}
	}

	if len(recentDurations) == 0 {
		return trini.DefaultFamily
	}

	avgDuration := int64(0)
	for _, d := range recentDurations {
		avgDuration += d
	}
	avgDuration /= int64(len(recentDurations))

	// Find best matching family
	for _, family := range trini.ProgramFamilies {
		if family.ID == "default" {
			continue // Skip default family in selection
		}

		criteria := family.EvaluationCriteria
		minSamples, _ := criteria["min_samples"].(int)

		if len(recentDurations) < minSamples {
			continue
		}

		matches := true

		if maxDuration, exists := criteria["max_magc_duration"].(int); exists {
			if avgDuration > int64(maxDuration) {
				matches = false
			}
		}

		if minDuration, exists := criteria["min_magc_duration"].(int); exists {
			if avgDuration < int64(minDuration) {
				matches = false
			}
		}

		if matches {
			return family
		}
	}

	return trini.DefaultFamily
}

// generateMaGCForecast implements the MaGA algorithm for MaGC prediction
func (s *Server) generateMaGCForecast(history []GCSnapshot) *MaGCForecast {
	if len(history) < 5 {
		return nil // Need minimum samples for forecasting
	}

	s.mu.Lock()
	family := s.CurrentFamily
	s.mu.Unlock()

	if family == nil {
		return nil
	}

	windowSize := family.ForecastWindowSize
	if windowSize > len(history) {
		windowSize = len(history)
	}

	// Get recent history window
	recentHistory := history[len(history)-windowSize:]

	// Step 1: Forecast YoungGen threshold when OldGen exhaustion occurs
	youngGenThreshold := s.forecastYoungGenThreshold(recentHistory)
	if youngGenThreshold <= 0 {
		return nil
	}

	// Step 2: Forecast time when YoungGen reaches threshold
	timeToMaGC := s.forecastTimeToMaGC(recentHistory, youngGenThreshold)
	if timeToMaGC <= 0 {
		return nil
	}

	// Calculate confidence based on data quality
	confidence := s.calculateForecastConfidence(recentHistory)

	return &MaGCForecast{
		PredictedTime:     time.Now().Add(time.Duration(timeToMaGC) * time.Millisecond),
		Confidence:        confidence,
		YoungGenThreshold: youngGenThreshold,
		TimeToMaGC:        timeToMaGC,
		ForecastCreatedAt: time.Now(),
	}
}

// forecastYoungGenThreshold predicts YoungGen memory when OldGen exhaustion occurs
func (s *Server) forecastYoungGenThreshold(history []GCSnapshot) int {
	if len(history) < 3 {
		return 0
	}

	// Linear regression: YoungGen = a * OldGen + b
	n := float64(len(history))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for _, snapshot := range history {
		x := float64(snapshot.OldGenUsed)
		y := float64(snapshot.YoungGenUsed)

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate regression coefficients
	denominator := n*sumX2 - sumX*sumX
	if math.Abs(denominator) < 1e-10 {
		return 0 // Avoid division by zero
	}

	a := (n*sumXY - sumX*sumY) / denominator
	b := (sumY - a*sumX) / n

	// Predict YoungGen when OldGen reaches 90% capacity
	oldGenThreshold := float64(s.OldGenMax) * 0.9
	youngGenThreshold := a*oldGenThreshold + b

	if youngGenThreshold < 0 {
		youngGenThreshold = 0
	}

	return int(youngGenThreshold)
}

// forecastTimeToMaGC predicts when YoungGen will reach the threshold
func (s *Server) forecastTimeToMaGC(history []GCSnapshot, youngGenThreshold int) int64 {
	if len(history) < 3 {
		return 0
	}

	// Linear regression: Time = a * YoungGenUsed + b
	n := float64(len(history))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	baseTime := history[0].Timestamp

	for _, snapshot := range history {
		x := float64(snapshot.YoungGenUsed)
		y := float64(snapshot.Timestamp.Sub(baseTime).Milliseconds())

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate regression coefficients
	denominator := n*sumX2 - sumX*sumX
	if math.Abs(denominator) < 1e-10 {
		return 0
	}

	a := (n*sumXY - sumX*sumY) / denominator
	b := (sumY - a*sumX) / n

	// Predict time when YoungGen reaches threshold
	predictedTime := a*float64(youngGenThreshold) + b
	currentTime := float64(time.Now().Sub(baseTime).Milliseconds())

	timeToMaGC := int64(predictedTime - currentTime)

	if timeToMaGC < 0 {
		return 0
	}

	return timeToMaGC
}

// calculateForecastConfidence calculates confidence based on data consistency
func (s *Server) calculateForecastConfidence(history []GCSnapshot) float64 {
	if len(history) < 3 {
		return 0.0
	}

	// Simple confidence based on data points and recency
	baseConfidence := math.Min(float64(len(history))/20.0, 1.0) // More data = higher confidence

	// Reduce confidence if data is old
	latestSnapshot := history[len(history)-1]
	timeSinceLatest := time.Since(latestSnapshot.Timestamp)
	if timeSinceLatest > 30*time.Second {
		baseConfidence *= 0.5
	}

	return baseConfidence
}

// IsMaGCPredicted checks if a MaGC is predicted within the threshold
func (s *Server) IsMaGCPredicted(thresholdMs int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.LastMaGCForecast == nil {
		return false
	}

	// Check if forecast is still valid (not too old)
	if time.Since(s.LastMaGCForecast.ForecastCreatedAt) > 30*time.Second {
		return false
	}

	// Check if MaGC is predicted within threshold
	timeToMaGC := time.Until(s.LastMaGCForecast.PredictedTime).Milliseconds()

	return timeToMaGC >= 0 && timeToMaGC <= thresholdMs
}
