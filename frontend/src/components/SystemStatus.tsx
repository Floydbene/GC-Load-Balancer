import React, { useEffect } from "react";
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
} from "@mui/material";
import {
  Refresh as RefreshIcon,
  Computer as ComputerIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
  Warning as WarningIcon,
  Memory as MemoryIcon,
} from "@mui/icons-material";
import { useTaskStore, type ServerStatus } from "../store/useTaskStore";

const ServerCard: React.FC<{ server: ServerStatus }> = ({ server }) => {
  const getServerStatusColor = (
    isAvailable: boolean,
    isCollectingGC: boolean
  ) => {
    if (isCollectingGC) return "warning";
    if (isAvailable) return "success";
    return "error";
  };

  const getServerStatusIcon = (
    isAvailable: boolean,
    isCollectingGC: boolean
  ) => {
    if (isCollectingGC) return <WarningIcon />;
    if (isAvailable) return <CheckCircleIcon />;
    return <CancelIcon />;
  };

  const parseMemoryUsage = (memUsage: string) => {
    const match = memUsage.match(/(\d+\.?\d*)%/);
    return match ? parseFloat(match[1]) : 0;
  };

  const memoryPercentage = parseMemoryUsage(server.mem_used);

  return (
    <Card variant="outlined" className="p-[10px]">
      <CardContent>
        <Box className="flex items-center justify-between mb-3">
          <Box className="flex items-center gap-[2] mb-[10px]">
            <Avatar
              sx={{
                bgcolor: "primary.main",
                width: 32,
                height: 32,
                margin: "0 10px",
              }}
            >
              <ComputerIcon fontSize="small" />
            </Avatar>
            <Typography variant="h6">Server {server.server_id}</Typography>
          </Box>
          <Chip
            icon={getServerStatusIcon(
              server.is_available,
              server.is_collecting_gc
            )}
            label={
              server.is_available
                ? server.is_collecting_gc
                  ? "GC"
                  : "Online"
                : "Busy"
            }
            color={
              getServerStatusColor(
                server.is_available,
                server.is_collecting_gc
              ) as any
            }
            size="small"
          />
        </Box>

        <Stack spacing={2}>
          <Box>
            <Box className="flex items-center justify-between mb-1">
              <Typography variant="body2" color="text.secondary">
                Status:
              </Typography>
              <Typography variant="body2" fontWeight="medium">
                {server.status}
              </Typography>
            </Box>
          </Box>

          <Box>
            <Box className="flex items-center justify-between mb-1">
              <Typography variant="body2" color="text.secondary">
                Available:
              </Typography>
              <Typography variant="body2" fontWeight="medium">
                {server.is_available ? "Yes" : "No"}
              </Typography>
            </Box>
          </Box>

          <Box>
            <Box className="flex items-center justify-between mb-1">
              <Typography variant="body2" color="text.secondary">
                Memory Usage:
              </Typography>
              <Typography variant="body2" fontWeight="medium">
                {server.mem_used}
              </Typography>
            </Box>
            <LinearProgress
              variant="determinate"
              value={memoryPercentage}
              color={
                memoryPercentage > 80
                  ? "error"
                  : memoryPercentage > 60
                  ? "warning"
                  : "primary"
              }
              sx={{ height: 6, borderRadius: 3 }}
            />
          </Box>

          <Box>
            <Box className="flex items-center justify-between">
              <Typography variant="body2" color="text.secondary">
                Tasks Processed:
              </Typography>
              <Typography variant="body2" fontWeight="medium">
                {server.tasks_processed}
              </Typography>
            </Box>
          </Box>
        </Stack>

        {server.is_collecting_gc && (
          <Box className="mt-3">
            <Chip
              icon={<MemoryIcon />}
              label="Collecting garbage"
              color="warning"
              variant="outlined"
              size="small"
            />
          </Box>
        )}
      </CardContent>
    </Card>
  );
};

export const SystemStatus: React.FC = () => {
  const { systemStatus, getSystemStatus, isLoading } = useTaskStore();

  useEffect(() => {
    getSystemStatus();
    const interval = setInterval(getSystemStatus, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, [getSystemStatus]);

  if (isLoading && !systemStatus) {
    return (
      <Box className="text-center py-8">
        <CircularProgress size={40} />
        <Typography variant="body1" color="text.secondary" className="mt-4">
          Loading system status...
        </Typography>
      </Box>
    );
  }

  if (!systemStatus) {
    return (
      <Box className="text-center py-4">
        <Typography variant="body1" color="error">
          Failed to load system status
        </Typography>
      </Box>
    );
  }

  return (
    <Box className="mb-8">
      <Box className="flex items-center justify-left mb-[20px] gap-[20px]">
        <Typography variant="h5" component="h2">
          System Status
        </Typography>
        <Button
          onClick={getSystemStatus}
          variant="outlined"
          size="small"
          startIcon={<RefreshIcon />}
          disabled={isLoading}
        >
          Refresh
        </Button>
      </Box>

      <Card variant="outlined" className="mb-4">
        <CardContent>
          <Box className="grid grid-cols-2 gap-4">
            <Box className="text-center">
              <Typography variant="h3" color="primary.main" fontWeight="bold">
                {systemStatus.total_servers}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Total Servers
              </Typography>
            </Box>
            <Box className="text-center">
              <Typography variant="h3" color="success.main" fontWeight="bold">
                {systemStatus.available_servers}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Available Servers
              </Typography>
            </Box>
          </Box>
        </CardContent>
      </Card>

      <Box className="grid grid-cols-2 md:grid-cols-2">
        {systemStatus.servers.map((server) => (
          <ServerCard key={server.server_id} server={server} />
        ))}
      </Box>
    </Box>
  );
};
