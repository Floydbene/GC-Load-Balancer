# TRINI GC-Aware Load Balancer Demo

This document demonstrates how to use the TRINI (GC-aware load balancing) implementation based on the research paper.

## Overview

TRINI implements the following key features from the paper:

1. **MaGA Algorithm**: Forecasts Major GC (MaGC) events using linear regression
2. **Program Families**: Classifies applications based on GC characteristics
3. **GC-Aware Algorithms**: Modified RR, RAN, WRR, WRAN that avoid servers with predicted MaGC
4. **MAPE-K Loop**: Monitor, Analyze, Plan, Execute with Knowledge base
5. **Self-Adaptation**: Automatically adjusts policies based on GC behavior

## Getting Started

1. **Start the Load Balancer**:

   ```bash
   go run main.go
   ```

2. **TRINI starts automatically** and begins monitoring servers

3. **Available Commands**:
   - `task <text>` - Send tasks to be processed
   - `trini status` - Show TRINI status and server classifications
   - `trini policy RR 2000` - Set policy (algorithm, threshold in ms)
   - `trini on/off` - Enable/disable GC-aware load balancing

## Demo Scenarios

### Scenario 1: Basic GC-Aware Load Balancing

```
# Send some initial tasks to generate GC history
> task hello world 1
> task hello world 2
> task hello world 3

# Check TRINI status
> trini status

# Send more tasks to see GC-aware selection
> task large task with more content to trigger memory usage
```

### Scenario 2: Program Family Classification

```
# Generate different GC patterns by sending varying task sizes
> task small
> task medium sized task
> task large task with significant content to consume more memory and trigger different GC behavior

# Check how servers are classified into families
> trini status

# Servers should adapt to different program families based on their GC characteristics
```

### Scenario 3: MaGC Forecasting

```
# Send tasks to build up memory pressure
> task memory intensive task 1
> task memory intensive task 2
> task memory intensive task 3

# Check forecasting status
> trini status

# You should see MaGC predictions with confidence levels
```

### Scenario 4: Policy Adaptation

```
# Start with Round Robin
> trini policy RR 1000

# Send tasks and observe behavior
> task test 1
> task test 2

# Switch to Weighted Round Robin for better handling
> trini policy WRR 3000

# Send more tasks to see different behavior
> task test 3
> task test 4
```

### Scenario 5: Compare GC-Aware vs Regular Load Balancing

```
# Disable TRINI to see regular load balancing
> trini off
> task regular 1
> task regular 2

# Enable TRINI to see GC-aware load balancing
> trini on
> task gc-aware 1
> task gc-aware 2

# Compare the server selection patterns
```

## Expected Behavior

### Program Family Classification

Servers will be automatically classified into families based on their MaGC patterns:

- **Short MaGC Duration** (< 500ms): Uses RR with 1s threshold
- **Medium MaGC Duration** (500ms-2s): Uses WRR with 3s threshold
- **Long MaGC Duration** (> 2s): Uses WRR with 5s threshold

### GC-Aware Server Selection

When MaGC is predicted within the threshold:

- Server is skipped in load balancing
- Next available server is selected
- Console shows "Server X skipped: MaGC predicted within Yms"

### MaGC Forecasting

The MaGA algorithm:

1. Tracks YoungGen and OldGen memory usage
2. Uses linear regression to predict when OldGen exhaustion occurs
3. Forecasts the time until MaGC event
4. Provides confidence levels based on data quality

### Adaptive Behavior

TRINI continuously:

- Monitors GC patterns every 2 seconds
- Analyzes and adapts program families every 10 seconds
- Updates load balancing policies based on dominant server family
- Maintains forecast accuracy through continuous learning

## Key Metrics to Observe

1. **Server Classifications**: How servers are grouped into program families
2. **MaGC Predictions**: Forecast accuracy and confidence levels
3. **Load Balancing Decisions**: Which servers are selected/skipped
4. **Policy Adaptations**: How policies change based on GC patterns
5. **Performance Impact**: Reduced task rejections due to GC avoidance

## Advanced Usage

### Custom Program Families

You can extend the system by modifying the `initializeDefaultFamilies()` function in `TRINI.go` to add custom program families with different:

- Evaluation criteria
- Load balancing policies
- Forecast parameters
- MaGC thresholds

### Monitoring Integration

The system provides detailed logging and can be integrated with monitoring systems to track:

- GC prediction accuracy
- Load balancing effectiveness
- Server performance metrics
- Family classification stability

This implementation demonstrates the core concepts from the TRINI research paper, showing how GC-aware load balancing can improve cluster performance by avoiding servers during predicted Major GC events.
