import React, { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  Typography,
  Box,
  Button,
  Avatar,
  LinearProgress,
  Chip,
  CircularProgress,
  Stack,
  Switch,
  FormControlLabel,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Alert,
  AlertTitle,
  Tabs,
  Tab,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Link,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
} from "@mui/material";
import {
  Refresh as RefreshIcon,
  Computer as ComputerIcon,
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
  Memory as MemoryIcon,
  Timeline as TimelineIcon,
  Psychology as PsychologyIcon,
  Settings as SettingsIcon,
  Speed as SpeedIcon,
  Info as InfoIcon,
  Science as ScienceIcon,
  AutoGraph as AutoGraphIcon,
  School as SchoolIcon,
  Link as LinkIcon,
} from "@mui/icons-material";
import { useTaskStore } from "../store/useTaskStore";
import type {
  LoadBalancingPolicy,
  TRINIServerDetails,
} from "../store/useTaskStore";

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`trini-tabpanel-${index}`}
      aria-labelledby={`trini-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

const TRINIServerCard: React.FC<{ server: TRINIServerDetails }> = ({
  server,
}) => {
  const { getGCHistory, gcHistories } = useTaskStore();
  const [showGCHistory, setShowGCHistory] = useState(false);

  const handleShowHistory = async () => {
    if (!showGCHistory) {
      await getGCHistory(server.server_id, 5);
    }
    setShowGCHistory(!showGCHistory);
  };

  const youngGenPercentage =
    server.young_gen_max > 0
      ? (server.young_gen_used / server.young_gen_max) * 100
      : 0;
  const oldGenPercentage =
    server.old_gen_max > 0
      ? (server.old_gen_used / server.old_gen_max) * 100
      : 0;

  return (
    <Card variant="outlined" sx={{ height: "100%", mb: 2 }}>
      <CardContent>
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            mb: 2,
          }}
        >
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <Avatar sx={{ bgcolor: "primary.main", width: 32, height: 32 }}>
              <ComputerIcon fontSize="small" />
            </Avatar>
            <Typography variant="h6">Server {server.server_id}</Typography>
          </Box>
          <Tooltip
            title={`Program Family: ${
              server.current_family?.name || "Unclassified"
            }. ${
              server.last_magc_forecast?.is_predicted_within_threshold
                ? "⚠️ MaGC predicted within threshold"
                : "✅ No imminent MaGC predicted"
            }`}
          >
            <Chip
              icon={
                server.last_magc_forecast?.is_predicted_within_threshold ? (
                  <WarningIcon />
                ) : (
                  <CheckCircleIcon />
                )
              }
              label={server.current_family?.name || "Unclassified"}
              color={
                server.last_magc_forecast?.is_predicted_within_threshold
                  ? "warning"
                  : "success"
              }
              size="small"
            />
          </Tooltip>
        </Box>

        <Stack spacing={2}>
          <Box>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Young Generation ({youngGenPercentage.toFixed(1)}%):
            </Typography>
            <LinearProgress
              variant="determinate"
              value={youngGenPercentage}
              color={
                youngGenPercentage > 80
                  ? "error"
                  : youngGenPercentage > 60
                  ? "warning"
                  : "primary"
              }
              sx={{ height: 6, borderRadius: 3 }}
            />
          </Box>

          <Box>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Old Generation ({oldGenPercentage.toFixed(1)}%):
            </Typography>
            <LinearProgress
              variant="determinate"
              value={oldGenPercentage}
              color={
                oldGenPercentage > 80
                  ? "error"
                  : oldGenPercentage > 60
                  ? "warning"
                  : "success"
              }
              sx={{ height: 6, borderRadius: 3 }}
            />
          </Box>

          {server.last_magc_forecast && (
            <Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                MaGC Forecast:
              </Typography>
              <Typography variant="caption" display="block">
                Time to MaGC:{" "}
                <strong>{server.last_magc_forecast.time_to_magc_ms}ms</strong>
              </Typography>
              <Typography variant="caption" display="block">
                Confidence:{" "}
                <strong>
                  {(server.last_magc_forecast.confidence * 100).toFixed(1)}%
                </strong>
              </Typography>
            </Box>
          )}

          <Button
            size="small"
            variant="outlined"
            startIcon={<TimelineIcon />}
            onClick={handleShowHistory}
            disabled={server.gc_history_count === 0}
          >
            {showGCHistory ? "Hide" : "Show"} GC History
          </Button>

          {showGCHistory && gcHistories[server.server_id] && (
            <TableContainer component={Paper} sx={{ maxHeight: 200 }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Time</TableCell>
                    <TableCell>Young Gen</TableCell>
                    <TableCell>Old Gen</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {gcHistories[server.server_id].gc_history
                    .slice(0, 3)
                    .map((snapshot, index) => (
                      <TableRow key={index}>
                        <TableCell>
                          <Typography variant="caption">
                            {new Date(snapshot.timestamp).toLocaleTimeString()}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Typography variant="caption">
                            {snapshot.young_gen_used}/{snapshot.young_gen_max}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Typography variant="caption">
                            {snapshot.old_gen_used}/{snapshot.old_gen_max}
                          </Typography>
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Stack>
      </CardContent>
    </Card>
  );
};

const PolicyControlPanel: React.FC = () => {
  const {
    triniStatus,
    updateTRINIPolicy,
    toggleTRINI,
    isTRINILoading,
    triniError,
  } = useTaskStore();
  const [policy, setPolicy] = useState<LoadBalancingPolicy>({
    algorithm: "RR",
    gc_aware: true,
    magc_threshold_ms: 2000,
    history_window_size: 30,
  });

  useEffect(() => {
    if (triniStatus?.current_policy) {
      setPolicy(triniStatus.current_policy);
    }
  }, [triniStatus]);

  const handlePolicyUpdate = async () => {
    await updateTRINIPolicy(policy);
  };

  const handleToggleTRINI = async (active: boolean) => {
    await toggleTRINI(active);
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Policy Control
        </Typography>

        <Stack spacing={3}>
          <Tooltip title="Toggle TRINI system on/off. When enabled, load balancing uses GC forecasts to avoid servers about to undergo Major GC events">
            <FormControlLabel
              control={
                <Switch
                  checked={triniStatus?.active || false}
                  onChange={(e) => handleToggleTRINI(e.target.checked)}
                  disabled={isTRINILoading}
                />
              }
              label="Enable TRINI GC-Aware Load Balancing"
            />
          </Tooltip>

          <Tooltip title="Select the load balancing algorithm. GC-aware versions (GC-RR, GC-RAN, GC-WRR, GC-WRAN) will avoid servers with predicted MaGC events">
            <FormControl fullWidth>
              <InputLabel>Algorithm</InputLabel>
              <Select
                value={policy.algorithm}
                label="Algorithm"
                onChange={(e) =>
                  setPolicy({ ...policy, algorithm: e.target.value })
                }
              >
                <MenuItem value="RR">Round Robin (RR)</MenuItem>
                <MenuItem value="RAN">Random (RAN)</MenuItem>
                <MenuItem value="WRR">Weighted Round Robin (WRR)</MenuItem>
                <MenuItem value="WRAN">Weighted Random (WRAN)</MenuItem>
              </Select>
            </FormControl>
          </Tooltip>

          <Tooltip title="Time threshold for MaGC avoidance. Servers with MaGC predicted within this time window will be skipped. Lower values = more aggressive avoidance">
            <FormControl fullWidth>
              <InputLabel>MaGC Threshold</InputLabel>
              <Select
                value={policy.magc_threshold_ms}
                label="MaGC Threshold"
                onChange={(e) =>
                  setPolicy({
                    ...policy,
                    magc_threshold_ms: e.target.value as number,
                  })
                }
              >
                <MenuItem value={500}>500ms - Very Aggressive</MenuItem>
                <MenuItem value={1000}>1 second - Aggressive</MenuItem>
                <MenuItem value={2000}>2 seconds - Balanced</MenuItem>
                <MenuItem value={5000}>5 seconds - Conservative</MenuItem>
                <MenuItem value={10000}>
                  10 seconds - Very Conservative
                </MenuItem>
                <MenuItem value={15000}>15 seconds - Relaxed</MenuItem>
                <MenuItem value={30000}>30 seconds - Minimal Impact</MenuItem>
              </Select>
            </FormControl>
          </Tooltip>

          <Button
            variant="contained"
            onClick={handlePolicyUpdate}
            disabled={isTRINILoading}
            startIcon={
              isTRINILoading ? <CircularProgress size={20} /> : <SettingsIcon />
            }
          >
            Update Policy
          </Button>

          {triniError && (
            <Alert severity="error">
              <AlertTitle>Policy Update Error</AlertTitle>
              {triniError}
            </Alert>
          )}
        </Stack>
      </CardContent>
    </Card>
  );
};

const AboutTRINI: React.FC = () => {
  return (
    <Stack spacing={4}>
      {/* Header */}
      <Card>
        <CardContent>
          <Box sx={{ display: "flex", alignItems: "center", gap: 2, mb: 2 }}>
            <Avatar sx={{ bgcolor: "primary.main" }}>
              <ScienceIcon />
            </Avatar>
            <Box>
              <Typography variant="h5" component="h2">
                TRINI: GC-Aware Load Balancing
              </Typography>
              <Typography variant="subtitle1" color="text.secondary">
                A self-adaptive load balancing strategy for garbage collection
                optimization
              </Typography>
            </Box>
          </Box>

          <Typography variant="body1" paragraph>
            TRINI (Threshold-based Resource-aware Intelligent Network
            Infrastructure) is a GC-aware load balancing strategy that
            dynamically adjusts to the specific garbage collection
            characteristics of applications. It forecasts Major GC (MaGC) events
            and uses this information to improve cluster performance by avoiding
            servers about to undergo GC pauses.
          </Typography>

          <Box sx={{ mt: 2 }}>
            <Button
              variant="outlined"
              startIcon={<LinkIcon />}
              href="https://onlinelibrary.wiley.com/doi/abs/10.1002/spe.2391"
              target="_blank"
              rel="noopener noreferrer"
            >
              Read the Research Paper
            </Button>
          </Box>
        </CardContent>
      </Card>

      {/* Load Balancing Algorithms */}
      <Card>
        <CardContent>
          <Typography
            variant="h6"
            gutterBottom
            sx={{ display: "flex", alignItems: "center", gap: 1 }}
          >
            <SpeedIcon color="primary" />
            GC-Aware Load Balancing Algorithms
          </Typography>

          <Stack spacing={3}>
            <Box>
              <Typography
                variant="subtitle2"
                fontWeight="bold"
                color="primary.main"
              >
                GC-RR (GC-Aware Round Robin)
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Sequentially distributes requests while skipping servers with
                predicted MaGC events within the threshold.
              </Typography>
            </Box>

            <Box>
              <Typography
                variant="subtitle2"
                fontWeight="bold"
                color="primary.main"
              >
                GC-RAN (GC-Aware Random)
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Randomly selects servers while avoiding those with imminent MaGC
                predictions.
              </Typography>
            </Box>

            <Box>
              <Typography
                variant="subtitle2"
                fontWeight="bold"
                color="primary.main"
              >
                GC-WRR (GC-Aware Weighted Round Robin)
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Uses server weights for proportional load distribution while
                considering GC forecasts.
              </Typography>
            </Box>

            <Box>
              <Typography
                variant="subtitle2"
                fontWeight="bold"
                color="primary.main"
              >
                GC-WRAN (GC-Aware Weighted Random)
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Combines weighted selection with random distribution, avoiding
                servers with predicted MaGC events.
              </Typography>
            </Box>
          </Stack>

          <Alert severity="info" sx={{ mt: 2 }}>
            <AlertTitle>Escape Condition</AlertTitle>
            All GC-aware algorithms include an escape condition to prevent
            infinite loops when all servers have predicted MaGC events within
            the threshold.
          </Alert>
        </CardContent>
      </Card>

      {/* MaGA Algorithm */}
      <Card>
        <CardContent>
          <Typography
            variant="h6"
            gutterBottom
            sx={{ display: "flex", alignItems: "center", gap: 1 }}
          >
            <TimelineIcon color="primary" />
            MaGA: Major GC Forecast Algorithm
          </Typography>

          <Typography variant="body2" color="text.secondary" paragraph>
            The Major GC Forecast Algorithm (MaGA) predicts when MaGC events
            will occur using linear regression on generational heap memory
            allocation patterns.
          </Typography>

          <Typography variant="subtitle2" fontWeight="bold" gutterBottom>
            Two-Step Forecasting Process:
          </Typography>

          <List dense>
            <ListItem sx={{ pl: 2 }}>
              <ListItemText
                primary="1. YoungGen Threshold Prediction"
                secondary="Uses OldGen historical data to predict how much YoungGen allocation will trigger the next MaGC"
              />
            </ListItem>
            <ListItem sx={{ pl: 2 }}>
              <ListItemText
                primary="2. Time-to-MaGC Calculation"
                secondary="Uses YoungGen allocation rate to predict when the threshold will be reached"
              />
            </ListItem>
          </List>

          <Box sx={{ bgcolor: "grey.50", p: 2, borderRadius: 1, mt: 2 }}>
            <Typography variant="caption" fontWeight="bold">
              Key Parameters:
            </Typography>
            <Typography variant="body2" component="div">
              • <strong>Forecast Window Size (FWS):</strong> Amount of
              historical data used for predictions
              <br />• <strong>MaGC Threshold:</strong> Time window for
              considering servers "about to GC"
              <br />• <strong>Confidence Score:</strong> Statistical reliability
              of the forecast based on data quality
            </Typography>
          </Box>
        </CardContent>
      </Card>
    </Stack>
  );
};

const TRINIOverview: React.FC = () => {
  const { triniStatus } = useTaskStore();

  if (!triniStatus) {
    return (
      <Alert severity="info">
        <AlertTitle>TRINI Status</AlertTitle>
        TRINI status not available. Make sure the backend is running with TRINI
        enabled.
      </Alert>
    );
  }

  const serversWithForecasts = triniStatus.servers.filter(
    (s) => s.last_magc_forecast
  ).length;
  const serversWithPredictedMaGC = triniStatus.servers.filter(
    (s) => s.last_magc_forecast?.is_predicted_within_threshold
  ).length;

  return (
    <Box>
      {/* Status Cards */}
      <Box sx={{ display: "flex", flexWrap: "wrap", gap: 2, mb: 3 }}>
        <Tooltip title="Shows whether TRINI GC-aware load balancing is currently active">
          <Card sx={{ flex: "1 1 200px" }}>
            <CardContent sx={{ textAlign: "center" }}>
              <Avatar
                sx={{
                  bgcolor: triniStatus.active ? "success.main" : "error.main",
                  mx: "auto",
                  mb: 1,
                }}
              >
                <PsychologyIcon />
              </Avatar>
              <Typography variant="h6">TRINI Status</Typography>
              <Typography
                variant="h4"
                color={triniStatus.active ? "success.main" : "error.main"}
              >
                {triniStatus.active ? "Active" : "Inactive"}
              </Typography>
            </CardContent>
          </Card>
        </Tooltip>

        <Tooltip title="Current load balancing algorithm and whether it's using GC-aware selection">
          <Card sx={{ flex: "1 1 200px" }}>
            <CardContent sx={{ textAlign: "center" }}>
              <Avatar sx={{ bgcolor: "primary.main", mx: "auto", mb: 1 }}>
                <SpeedIcon />
              </Avatar>
              <Typography variant="h6">Algorithm</Typography>
              <Typography variant="h4" color="primary.main">
                {triniStatus.current_policy.algorithm}
              </Typography>
              <Typography variant="caption">
                {triniStatus.current_policy.gc_aware ? "GC-Aware" : "Regular"}
              </Typography>
            </CardContent>
          </Card>
        </Tooltip>

        <Tooltip title="Number of servers that have active MaGC forecasts from the MaGA algorithm">
          <Card sx={{ flex: "1 1 200px" }}>
            <CardContent sx={{ textAlign: "center" }}>
              <Avatar sx={{ bgcolor: "warning.main", mx: "auto", mb: 1 }}>
                <MemoryIcon />
              </Avatar>
              <Typography variant="h6">Forecasts</Typography>
              <Typography variant="h4" color="warning.main">
                {serversWithForecasts}
              </Typography>
              <Typography variant="caption">
                Servers with MaGC forecasts
              </Typography>
            </CardContent>
          </Card>
        </Tooltip>

        <Tooltip title="Number of servers predicted to have MaGC events within the current threshold - these servers are being avoided by GC-aware algorithms">
          <Card sx={{ flex: "1 1 200px" }}>
            <CardContent sx={{ textAlign: "center" }}>
              <Avatar
                sx={{
                  bgcolor:
                    serversWithPredictedMaGC > 0
                      ? "error.main"
                      : "success.main",
                  mx: "auto",
                  mb: 1,
                }}
              >
                <WarningIcon />
              </Avatar>
              <Typography variant="h6">Predicted MaGC</Typography>
              <Typography
                variant="h4"
                color={
                  serversWithPredictedMaGC > 0 ? "error.main" : "success.main"
                }
              >
                {serversWithPredictedMaGC}
              </Typography>
              <Typography variant="caption">Within threshold</Typography>
            </CardContent>
          </Card>
        </Tooltip>
      </Box>

      {/* Current Policy */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Current Policy
          </Typography>
          <Stack spacing={1}>
            <Box sx={{ display: "flex", justifyContent: "space-between" }}>
              <Typography variant="body2">Algorithm:</Typography>
              <Typography variant="body2" fontWeight="bold">
                {triniStatus.current_policy.algorithm}
              </Typography>
            </Box>
            <Box sx={{ display: "flex", justifyContent: "space-between" }}>
              <Typography variant="body2">GC-Aware:</Typography>
              <Typography variant="body2" fontWeight="bold">
                {triniStatus.current_policy.gc_aware ? "Enabled" : "Disabled"}
              </Typography>
            </Box>
            <Box sx={{ display: "flex", justifyContent: "space-between" }}>
              <Typography variant="body2">MaGC Threshold:</Typography>
              <Typography variant="body2" fontWeight="bold">
                {triniStatus.current_policy.magc_threshold_ms}ms
              </Typography>
            </Box>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};

export const TRINIDashboard: React.FC = () => {
  const {
    triniStatus,
    getTRINIStatus,
    getProgramFamilies,
    isTRINILoading,
    triniError,
    clearTRINIError,
  } = useTaskStore();

  const [tabValue, setTabValue] = useState(0);

  useEffect(() => {
    const fetchData = async () => {
      await getTRINIStatus();
      await getProgramFamilies();
    };

    fetchData();

    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);
  }, [getTRINIStatus, getProgramFamilies]);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleRefresh = async () => {
    clearTRINIError();
    await getTRINIStatus();
    await getProgramFamilies();
  };

  if (isTRINILoading && !triniStatus) {
    return (
      <Box sx={{ textAlign: "center", py: 8 }}>
        <CircularProgress size={60} />
        <Typography variant="h6" sx={{ mt: 2 }} color="text.secondary">
          Loading TRINI Dashboard...
        </Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          mb: 4,
        }}
      >
        <Typography variant="h4" component="h1">
          🔍 TRINI GC-Aware Load Balancer
        </Typography>
        <Button
          onClick={handleRefresh}
          variant="outlined"
          startIcon={
            isTRINILoading ? <CircularProgress size={20} /> : <RefreshIcon />
          }
          disabled={isTRINILoading}
        >
          Refresh
        </Button>
      </Box>

      <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 3 }}>
        <Tabs value={tabValue} onChange={handleTabChange}>
          <Tab label="Overview" />
          <Tab label="Server Details" />
          <Tab label="Policy Control" />
          <Tab
            label={
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <InfoIcon fontSize="small" />
                About TRINI
              </Box>
            }
          />
        </Tabs>
      </Box>

      <TabPanel value={tabValue} index={0}>
        <TRINIOverview />
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          {triniStatus?.servers.map((server) => (
            <TRINIServerCard key={server.server_id} server={server} />
          ))}
        </Box>
      </TabPanel>

      <TabPanel value={tabValue} index={2}>
        <PolicyControlPanel />
      </TabPanel>

      <TabPanel value={tabValue} index={3}>
        <AboutTRINI />
      </TabPanel>
    </Box>
  );
};
