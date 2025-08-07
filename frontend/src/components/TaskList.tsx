import React from "react";
import {
  Card,
  CardContent,
  Typography,
  Chip,
  Box,
  Button,
  Stack,
  Divider,
  type SxProps,
} from "@mui/material";
import {
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
  Schedule as ScheduleIcon,
  Clear as ClearIcon,
} from "@mui/icons-material";
import { useTaskStore, type Task } from "../store/useTaskStore";

const TaskItem: React.FC<{ task: Task }> = ({ task }) => {
  const getStatusColor = (status: Task["status"]) => {
    switch (status) {
      case "completed":
        return "success";
      case "rejected":
        return "error";
      case "pending":
        return "warning";
      default:
        return "default";
    }
  };

  const getStatusIcon = (status: Task["status"]) => {
    switch (status) {
      case "completed":
        return <CheckCircleIcon fontSize="small" />;
      case "rejected":
        return <CancelIcon fontSize="small" />;
      case "pending":
        return <ScheduleIcon fontSize="small" />;
      default:
        return null;
    }
  };

  return (
    <Card variant="outlined" className="mb-4">
      <CardContent>
        <Box className="flex items-center justify-between mb-2">
          <Chip
            icon={getStatusIcon(task.status) || undefined}
            label={task.status.toUpperCase()}
            color={getStatusColor(task.status) as any}
            size="small"
          />
          <Typography variant="caption" color="text.secondary">
            {new Date(task.created_at).toLocaleTimeString()}
          </Typography>
        </Box>

        <Box className="mb-2">
          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
            Input:
          </Typography>
          <Typography variant="body1" fontWeight="medium">
            {task.input}
          </Typography>
        </Box>

        {task.status === "rejected" && task.output && (
          <Box className="mb-2">
            <Typography variant="subtitle2" color="error" gutterBottom>
              Error:
            </Typography>
            <Box>
              <Typography variant="body2">{task.output}</Typography>
            </Box>
          </Box>
        )}

        {task.output && (
          <Box className="mb-2">
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              Output:
            </Typography>
            <Box className="bg-gray-50 p-2 rounded">
              <Typography
                variant="body2"
                fontFamily="monospace"
                sx={{ overflow: "scroll" }}
              >
                {task.output}
              </Typography>
            </Box>
          </Box>
        )}

        <Divider className="my-2" />
        <Typography variant="caption" color="text.disabled">
          ID: {task.id}
        </Typography>
      </CardContent>
    </Card>
  );
};

export const TaskList: React.FC<{ sx?: SxProps }> = ({ sx }) => {
  const { tasks, clearTasks } = useTaskStore();

  if (tasks.length === 0) {
    return (
      <Box className="text-center py-8">
        <Typography variant="body1" color="text.secondary">
          No tasks submitted yet. Submit a task above to get started!
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ ...sx, overflow: "scroll", height: "30vh" }}>
      <Box className="flex items-center justify-between mb-[20px]">
        <Typography variant="h5" component="h2">
          Tasks ({tasks.length})
        </Typography>
        <Button
          onClick={clearTasks}
          variant="outlined"
          size="small"
          startIcon={<ClearIcon />}
          color="inherit"
        >
          Clear All
        </Button>
      </Box>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr 1fr" },
          gap: 2,
        }}
      >
        {tasks.map((task) => (
          <TaskItem key={task.id} task={task} />
        ))}
      </Box>
    </Box>
  );
};
