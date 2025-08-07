package main

import (
	"bufio"
	"fmt"
	"golang_lb/server"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	lb := &server.LoadBalancer{
		Servers: make([]*server.Server, 0),
	}

	for i := 1; i <= 3; i++ {
		srv := &server.Server{
			ID:           i,
			LoadBalancer: lb,
			TaskStorage:  make([]string, 0),
		}
		lb.Servers = append(lb.Servers, srv)
	}

	// Start the load balancer
	fmt.Println("🚀 Starting Load Balancer System...")
	lb.Start()
	time.Sleep(100 * time.Millisecond)

	// Start TRINI GC-aware load balancing
	fmt.Println("🔍 Starting TRINI GC-aware load balancing...")
	lb.StartTRINI()
	time.Sleep(500 * time.Millisecond)

	fmt.Println("✅ Load Balancer with TRINI is ready!")
	printTRINIHelp()

	// Interactive command loop
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		switch command {
		case "task", "t":
			if len(parts) < 2 {
				fmt.Println("❌ Usage: task <your_task_string>")
				continue
			}
			taskInput := strings.Join(parts[1:], " ")
			handleTask(lb, taskInput)

		case "ping", "p":
			if len(parts) < 2 {
				fmt.Println("❌ Usage: ping <server_id>")
				continue
			}
			serverID, err := strconv.Atoi(parts[1])
			if err != nil || serverID < 1 || serverID > len(lb.Servers) {
				fmt.Printf("❌ Invalid server ID. Use 1-%d\n", len(lb.Servers))
				continue
			}
			handlePing(lb, serverID)

		case "status", "s":
			handleStatus(lb)

		case "trini":
			if len(parts) < 2 {
				fmt.Println("❌ Usage: trini <on|off|status|policy>")
				continue
			}
			handleTRINI(lb, parts[1:])

		case "help", "h":
			printTRINIHelp()

		case "quit", "q", "exit":
			fmt.Println("👋 Goodbye!")
			return

		default:
			fmt.Printf("❌ Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

func printTRINIHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  task <text>     - Send a task to be processed (alias: t)")
	fmt.Println("  ping <id>       - Ping a specific server (alias: p)")
	fmt.Println("  status          - Show all servers status (alias: s)")
	fmt.Println("  trini <cmd>     - TRINI GC-aware control (on|off|status|policy)")
	fmt.Println("  help            - Show this help message (alias: h)")
	fmt.Println("  quit            - Exit the program (alias: q, exit)")
	fmt.Println("\nExample: task hello world")
	fmt.Println("Example: ping 1")
	fmt.Println("Example: trini status")
}

func handleTRINI(lb *server.LoadBalancer, args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: trini <on|off|status|policy>")
		return
	}

	command := strings.ToLower(args[0])

	switch command {
	case "on":
		if lb.TRINI != nil {
			lb.TRINI.IsActive = true
			fmt.Println("✅ TRINI GC-aware load balancing enabled")
		} else {
			fmt.Println("❌ TRINI not initialized")
		}

	case "off":
		if lb.TRINI != nil {
			lb.TRINI.IsActive = false
			fmt.Println("⚠️ TRINI GC-aware load balancing disabled")
		} else {
			fmt.Println("❌ TRINI not initialized")
		}

	case "status":
		showTRINIStatus(lb)

	case "policy":
		if len(args) > 1 {
			setPolicyFromArgs(lb, args[1:])
		} else {
			showCurrentPolicy(lb)
		}

	default:
		fmt.Printf("❌ Unknown TRINI command: %s\n", command)
		fmt.Println("Available: on, off, status, policy")
	}
}

func showTRINIStatus(lb *server.LoadBalancer) {
	if lb.TRINI == nil {
		fmt.Println("❌ TRINI not initialized")
		return
	}

	fmt.Println("\n🔍 TRINI Status:")
	fmt.Printf("   Active: %t\n", lb.TRINI.IsActive)
	fmt.Printf("   Monitor Interval: %v\n", lb.TRINI.MonitorInterval)
	fmt.Printf("   Analysis Interval: %v\n", lb.TRINI.AnalysisInterval)
	fmt.Printf("   Program Families: %d\n", len(lb.TRINI.ProgramFamilies))

	fmt.Println("\n📊 Server Family Classifications:")
	for _, server := range lb.Servers {
		if server.CurrentFamily != nil {
			fmt.Printf("   Server %d: %s\n", server.ID, server.CurrentFamily.Name)
			if server.LastMaGCForecast != nil {
				timeToMaGC := server.LastMaGCForecast.TimeToMaGC
				fmt.Printf("     Next MaGC in: %dms (confidence: %.2f)\n",
					timeToMaGC, server.LastMaGCForecast.Confidence)
			}
		} else {
			fmt.Printf("   Server %d: Not classified\n", server.ID)
		}
	}

	fmt.Println("\n🔧 Current Policy:")
	fmt.Printf("   Algorithm: %s\n", lb.CurrentPolicy.Algorithm)
	fmt.Printf("   GC-Aware: %t\n", lb.CurrentPolicy.GCAware)
	fmt.Printf("   MaGC Threshold: %dms\n", lb.CurrentPolicy.MaGCThreshold)
}

func showCurrentPolicy(lb *server.LoadBalancer) {
	fmt.Println("\n🔧 Current Load Balancing Policy:")
	fmt.Printf("   Algorithm: %s\n", lb.CurrentPolicy.Algorithm)
	fmt.Printf("   GC-Aware: %t\n", lb.CurrentPolicy.GCAware)
	fmt.Printf("   MaGC Threshold: %dms\n", lb.CurrentPolicy.MaGCThreshold)
	fmt.Printf("   History Window: %d\n", lb.CurrentPolicy.HistoryWindowSize)
}

func setPolicyFromArgs(lb *server.LoadBalancer, args []string) {
	if len(args) < 2 {
		fmt.Println("❌ Usage: trini policy <algorithm> <threshold_ms>")
		fmt.Println("Algorithms: RR, RAN, WRR, WRAN")
		return
	}

	algorithm := strings.ToUpper(args[0])
	threshold, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Println("❌ Invalid threshold value")
		return
	}

	policy := server.LoadBalancingPolicy{
		Algorithm:         algorithm,
		GCAware:           true,
		MaGCThreshold:     threshold,
		HistoryWindowSize: 30,
	}

	lb.SetLoadBalancingPolicy(policy)
}

func handleTask(lb *server.LoadBalancer, taskInput string) {
	fmt.Printf("📤 Sending task: '%s'\n", taskInput)

	srv := lb.GetServerForTask(taskInput)
	if srv != nil {
		response := srv.RequestTask(taskInput)
		fmt.Printf("⏳ %s - %s\n", response.Status, response.Message)

		go func(resultChan chan *server.Task) {
			if result := <-resultChan; result != nil {
				if result.Status == "rejected" {
					fmt.Printf("\n❌ TASK REJECTED: '%s' (ID: %s) - Server overloaded\n> ",
						result.Input, result.ID)
				} else {
					fmt.Printf("\n🎉 TASK COMPLETED: '%s' → '%s' (ID: %s)\n> ",
						result.Input, result.Output, result.ID)
				}
			}
		}(response.ResultChan)
	} else {
		fmt.Println("❌ No available server found!")
	}
}

func handlePing(lb *server.LoadBalancer, serverID int) {
	server := lb.Servers[serverID-1]
	pingResult := server.Ping()

	fmt.Printf("🏓 Ping Server %d:\n", serverID)
	fmt.Printf("   Status: %s\n", pingResult["status"])
	fmt.Printf("   Available: %v\n", pingResult["is_available"])
	fmt.Printf("   Memory Usage: %v\n", pingResult["mem_used"])
	fmt.Printf("   Collecting GC: %v\n", pingResult["is_collecting_gc"])
	fmt.Printf("   Tasks Processed: %d\n", pingResult["tasks_processed"])
	if taskIDs, ok := pingResult["task_ids"].([]string); ok && len(taskIDs) > 0 {
		fmt.Printf("   Recent Task IDs: %v\n", taskIDs)
	}
}

func handleStatus(lb *server.LoadBalancer) {
	fmt.Println("📊 Load Balancer Status:")
	fmt.Printf("   Total Servers: %d\n", len(lb.Servers))

	availableCount := 0
	for _, server := range lb.Servers {
		pingResult := server.Ping()
		isAvailable := pingResult["is_available"].(bool)
		if isAvailable {
			availableCount++
		}

		status := "🟢 Available"
		if !isAvailable {
			status = "🔴 Busy"
		}
		if pingResult["is_collecting_gc"].(bool) {
			status = "🟡 GC Mode"
		}

		fmt.Printf("   Server %d: %s (Tasks: %d)\n",
			server.ID, status, pingResult["tasks_processed"])
	}

	fmt.Printf("   Available Servers: %d/%d\n", availableCount, len(lb.Servers))
}
