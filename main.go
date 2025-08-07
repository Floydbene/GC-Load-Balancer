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

	fmt.Println("✅ Load Balancer is ready!")
	printHelp()

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

		case "help", "h":
			printHelp()

		case "quit", "q", "exit":
			fmt.Println("👋 Goodbye!")
			return

		default:
			fmt.Printf("❌ Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

func printHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  task <text>     - Send a task to be processed (alias: t)")
	fmt.Println("  ping <id>       - Ping a specific server (alias: p)")
	fmt.Println("  status          - Show all servers status (alias: s)")
	fmt.Println("  help            - Show this help message (alias: h)")
	fmt.Println("  quit            - Exit the program (alias: q, exit)")
	fmt.Println("\nExample: task hello world")
	fmt.Println("Example: ping 1")
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
