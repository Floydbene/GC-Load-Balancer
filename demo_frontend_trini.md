# TRINI Frontend Dashboard Demo

This guide shows how to use the enhanced frontend to monitor TRINI GC-aware load balancing in real-time.

## 🚀 Getting Started

1. **Start the Backend Server**:

   ```bash
   cd cmd/backend-server
   go run main.go middleware.go
   ```

   The server will start on `http://localhost:8080` with TRINI enabled.

2. **Start the Frontend**:
   ```bash
   cd frontend
   npm run dev
   ```
   The frontend will be available at `http://localhost:5173`

## 📊 Dashboard Features

### Navigation

The frontend now has two main views accessible via the top navigation:

- **Task Manager**: Original task submission and monitoring interface
- **TRINI Dashboard**: Comprehensive GC-aware load balancing monitoring

### TRINI Dashboard Tabs

#### 1. Overview Tab

- **System Status Cards**: Shows TRINI active/inactive, current algorithm, forecasts, and predicted MaGC events
- **Program Family Distribution**: Visual representation of how servers are classified
- **Current Policy Details**: All active policy settings and intervals

#### 2. Server Details Tab

- **Individual Server Cards**: Each server shows:
  - Current program family classification
  - Young and Old generation memory usage with progress bars
  - GC statistics (count, history entries, weights)
  - Real-time MaGC forecasts with confidence levels
  - Expandable GC history table with recent snapshots

#### 3. Policy Control Tab

- **Live Policy Management**:
  - Toggle TRINI on/off
  - Change load balancing algorithms (RR, RAN, WRR, WRAN)
  - Adjust MaGC threshold and history window size
  - Real-time policy updates
- **Program Families Reference**: Shows all available families and their criteria

## 🔍 Real-Time Monitoring Features

### Auto-Refresh

- Dashboard auto-refreshes every 5 seconds
- Manual refresh button available
- Real-time updates without page reload

### Visual Indicators

- **Color-coded server status**:
  - 🟢 Green: Healthy servers with proper classification
  - 🟡 Yellow: Servers with predicted MaGC events
  - 🔴 Red: Servers with issues or high memory usage
- **Progress bars** for memory usage (Young/Old generation)
- **Confidence indicators** for MaGC forecasts

### Interactive Elements

- **Expandable GC History**: Click to view detailed GC snapshots
- **Real-time Policy Updates**: Changes apply immediately
- **Error Handling**: Clear error messages and recovery options

## 📈 Demo Scenarios

### Scenario 1: Monitor Server Classification

1. Go to **TRINI Dashboard → Overview**
2. Submit several tasks of different sizes in the **Task Manager**
3. Watch the **Program Family Distribution** change as servers get classified
4. Observe how servers adapt from "Unclassified" to specific families

### Scenario 2: Watch GC Forecasting

1. Go to **TRINI Dashboard → Server Details**
2. Submit memory-intensive tasks to build up memory pressure
3. Watch the **MaGC Forecast** sections update with predictions
4. See confidence levels and time-to-MaGC estimates
5. Click **Show GC History** to see detailed snapshots

### Scenario 3: Policy Experimentation

1. Go to **TRINI Dashboard → Policy Control**
2. Try different algorithms:
   - Start with **Round Robin (RR)**
   - Switch to **Weighted Round Robin (WRR)**
   - Compare server selection patterns
3. Adjust the **MaGC Threshold**:
   - Lower values (1000ms) = more aggressive avoidance
   - Higher values (5000ms) = more tolerance
4. Watch how policy changes affect the **Overview** metrics

### Scenario 4: Compare GC-Aware vs Regular

1. **Enable TRINI** and submit tasks → observe server selection
2. **Disable TRINI** and submit similar tasks → compare behavior
3. Monitor the difference in:
   - Server utilization patterns
   - Task completion times
   - GC event timing

## 🎯 Key Metrics to Watch

### Overview Metrics

- **TRINI Status**: Active/Inactive indicator
- **Current Algorithm**: Which load balancing method is active
- **Forecasts**: How many servers have MaGC predictions
- **Predicted MaGC**: Servers expected to have GC events soon

### Server-Level Metrics

- **Program Family**: How each server is classified
- **Memory Usage**: Young/Old generation utilization
- **GC Statistics**: Event counts and patterns
- **Forecast Accuracy**: Confidence levels and timing

### Policy Effectiveness

- **Family Distribution**: How well servers are classified
- **Threshold Impact**: Effect of different MaGC thresholds
- **Algorithm Performance**: Comparison between different methods

## 💡 Tips for Effective Monitoring

1. **Use Multiple Browser Tabs**:

   - One for Task Manager (submitting tasks)
   - One for TRINI Dashboard (monitoring)

2. **Watch the Patterns**:

   - Server classifications evolve over time
   - GC forecasts become more accurate with more data
   - Different task sizes trigger different GC behaviors

3. **Experiment with Policies**:

   - Start conservative (high thresholds)
   - Gradually make more aggressive
   - Observe the impact on system performance

4. **Monitor Edge Cases**:
   - What happens when all servers predict MaGC?
   - How does the system recover after GC events?
   - How do different algorithms handle server failures?

## 🛠️ Troubleshooting

### Dashboard Not Loading

- Ensure backend server is running on port 8080
- Check browser console for API connection errors
- Verify TRINI is initialized in the backend logs

### No TRINI Data

- Confirm TRINI is active in the Policy Control tab
- Submit some tasks to generate GC activity
- Allow time for data collection (2-10 seconds)

### Forecast Not Appearing

- Servers need sufficient GC history (5+ samples)
- Memory usage must reach threshold levels
- Check server classifications are not "default"

The enhanced frontend provides complete visibility into the TRINI system, making it easy to understand and optimize GC-aware load balancing behavior!
