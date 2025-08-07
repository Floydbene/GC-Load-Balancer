import React, { useEffect, useState } from "react";
import {
  Paper,
  Typography,
  Box,
  Alert,
  AlertTitle,
  Button,
  CircularProgress,
  Fade,
} from "@mui/material";
import {
  Refresh as RefreshIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
  CheckCircle as CheckCircleIcon,
} from "@mui/icons-material";
import { useTaskStore } from "../store/useTaskStore";

interface TaskFormWrapperProps {
  children: React.ReactNode;
}

export const TaskFormWrapper: React.FC<TaskFormWrapperProps> = ({
  children,
}) => {
  const { systemStatus, error } = useTaskStore();
  const [isCheckingStatus, setIsCheckingStatus] = useState(false);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [previousStatus, setPreviousStatus] =
    useState<typeof systemStatus>(null);
  const [hasServerError, setHasServerError] = useState(false);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(
    null
  );

  // Custom status check function that handles 500 errors
  const checkStatusWithErrorHandling = async (showLoading: boolean = false) => {
    if (hasServerError) {
      return; // Stop checking if we've encountered a server error
    }

    if (showLoading) {
      setIsCheckingStatus(true);
    }

    try {
      const response = await fetch("/api/v1/status");

      if (response.status === 500) {
        // Stop polling on 500 error and maintain current state
        setHasServerError(true);
        if (pollingInterval) {
          clearInterval(pollingInterval);
          setPollingInterval(null);
        }
        if (showLoading) {
          setIsCheckingStatus(false);
        }
        return;
      }

      if (!response.ok) {
        throw new Error(
          `HTTP ${response.status}: Failed to fetch system status`
        );
      }

      const data = await response.json();

      // Update the store directly to avoid triggering other effects
      useTaskStore.setState({
        systemStatus: data,
        error: null,
      });

      setHasServerError(false); // Reset error state on successful fetch

      if (showLoading) {
        setIsCheckingStatus(false);
      }

      if (isInitialLoad) {
        setIsInitialLoad(false);
      }
    } catch (error) {
      // For non-500 errors, continue polling but set error state
      useTaskStore.setState({
        error:
          error instanceof Error ? error.message : "Unknown error occurred",
      });

      if (showLoading) {
        setIsCheckingStatus(false);
      }

      if (isInitialLoad) {
        setIsInitialLoad(false);
      }
    }
  };

  // Check system status on mount and set up periodic checks
  useEffect(() => {
    checkStatusWithErrorHandling(true);

    // Set up polling interval only if no server error
    if (!hasServerError) {
      const interval = setInterval(() => {
        checkStatusWithErrorHandling(false);
      }, 30000);

      setPollingInterval(interval);

      return () => {
        clearInterval(interval);
        setPollingInterval(null);
      };
    }
  }, [hasServerError]);

  // Track previous status to prevent flickering
  useEffect(() => {
    if (systemStatus) {
      setPreviousStatus(systemStatus);
    }
  }, [systemStatus]);

  // Cleanup polling interval on unmount
  useEffect(() => {
    return () => {
      if (pollingInterval) {
        clearInterval(pollingInterval);
      }
    };
  }, [pollingInterval]);

  const handleRefreshStatus = async () => {
    // Reset server error state to allow retry
    if (hasServerError) {
      setHasServerError(false);
    }
    await checkStatusWithErrorHandling(true);

    // Restart polling if it was stopped due to server error
    if (hasServerError === false && !pollingInterval) {
      const interval = setInterval(() => {
        checkStatusWithErrorHandling(false);
      }, 30000);
      setPollingInterval(interval);
    }
  };

  // Use current status or fallback to previous to prevent flickering
  const currentStatus = systemStatus || previousStatus;

  const isBackendAvailable = () => {
    if (hasServerError) return false;
    if (!currentStatus) return false;
    return currentStatus.available_servers > 0;
  };

  const getStatusMessage = () => {
    if (hasServerError) {
      return {
        severity: "error" as const,
        title: "Server Error Detected",
        message:
          "The backend returned a server error (500). Polling has been stopped to prevent further issues. Click refresh to retry.",
        icon: <ErrorIcon />,
      };
    }

    if (!currentStatus) {
      return {
        severity: "error" as const,
        title: "Backend Connection Failed",
        message:
          "Unable to connect to the load balancer backend. Please check your connection and try again.",
        icon: <ErrorIcon />,
      };
    }

    if (currentStatus.available_servers === 0) {
      return {
        severity: "warning" as const,
        title: "All Servers Busy",
        message: `All ${currentStatus.total_servers} servers are currently busy or collecting garbage. Please wait a moment and try again.`,
        icon: <WarningIcon />,
      };
    }

    if (currentStatus.available_servers < currentStatus.total_servers) {
      return {
        severity: "info" as const,
        title: "Limited Server Availability",
        message: `${currentStatus.available_servers} of ${currentStatus.total_servers} servers are available. Tasks may take longer to process.`,
        icon: <WarningIcon />,
      };
    }

    return {
      severity: "success" as const,
      title: "All Systems Operational",
      message: `All ${currentStatus.total_servers} servers are available and ready to process tasks.`,
      icon: <CheckCircleIcon />,
    };
  };

  const statusInfo = getStatusMessage();
  const canSubmitTasks = isBackendAvailable() && !error;

  if (isInitialLoad && isCheckingStatus && !currentStatus) {
    return (
      <Paper elevation={1} sx={{ p: 4, mb: 3, height: "fit-content" }}>
        <Box className="text-center py-8">
          <CircularProgress size={40} />
          <Typography variant="h6" className="mt-4" color="text.secondary">
            Checking backend status...
          </Typography>
          <Typography variant="body2" color="text.secondary" className="mt-2">
            Please wait while we verify the system is ready.
          </Typography>
        </Box>
      </Paper>
    );
  }

  return (
    <Paper elevation={1} sx={{ p: 3, mb: 3, height: "fit-content" }}>
      <Typography variant="h5" component="h2" gutterBottom>
        Submit Task
      </Typography>

      {/* Status Alert */}
      <Fade in={true} timeout={500}>
        <Alert
          severity={statusInfo.severity}
          icon={statusInfo.icon}
          sx={{ mb: 3 }}
          action={
            <Button
              color="inherit"
              size="small"
              onClick={handleRefreshStatus}
              disabled={isCheckingStatus}
              startIcon={
                isCheckingStatus ? (
                  <CircularProgress size={16} color="inherit" />
                ) : (
                  <RefreshIcon />
                )
              }
            >
              {isCheckingStatus ? "Checking..." : "Refresh"}
            </Button>
          }
        >
          <AlertTitle>{statusInfo.title}</AlertTitle>
          {statusInfo.message}
        </Alert>
      </Fade>

      {/* Task Form or Disabled State */}
      {canSubmitTasks ? (
        <Fade in={canSubmitTasks} timeout={800}>
          <Box>{children}</Box>
        </Fade>
      ) : (
        <Fade in={!canSubmitTasks} timeout={500}>
          <Box
            sx={{ opacity: 0.6, pointerEvents: "none", position: "relative" }}
          >
            {children}
            <Box
              sx={{
                position: "absolute",
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: "rgba(255, 255, 255, 0.8)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                borderRadius: 1,
              }}
            >
              <Typography variant="h6" color="text.secondary">
                Task submission temporarily disabled
              </Typography>
            </Box>
          </Box>
        </Fade>
      )}

      {/* Additional Error Display */}
      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          <AlertTitle>Connection Error</AlertTitle>
          {error}
        </Alert>
      )}
    </Paper>
  );
};
