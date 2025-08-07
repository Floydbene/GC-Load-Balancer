import React from "react";
import {
  Container,
  Typography,
  Box,
  AppBar,
  Toolbar,
  Paper,
  CssBaseline,
  ThemeProvider,
  createTheme,
} from "@mui/material";
import { TaskForm } from "./components/TaskForm";
import { TaskList } from "./components/TaskList";
import { SystemStatus } from "./components/SystemStatus";
import { ErrorAlert } from "./components/ErrorAlert";

const theme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: "#1976d2",
    },
    secondary: {
      main: "#dc004e",
    },
  },
});

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box className="min-h-screen bg-gray-50">
        <AppBar position="static" elevation={1}>
          <Toolbar sx={{ mx: "100px" }}>
            <Typography variant="h5" component="div" sx={{ flexGrow: 1 }}>
              🚀 Load Balancer Dashboard
            </Typography>
          </Toolbar>
        </AppBar>

        <Container maxWidth="xl" sx={{ py: 4 }}>
          <Box className="mb-[6px]">
            <Typography variant="body1" color="text.secondary">
              Submit tasks to the load balancer and monitor server status in
              real-time
            </Typography>
          </Box>

          <ErrorAlert />

          <Box className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <Box
              sx={{
                display: "flex",
                flexDirection: "row",
                gap: 2,
              }}
            >
              <Paper
                elevation={1}
                sx={{ p: 3, mb: 3, width: "50%", height: "fit-content" }}
              >
                <Typography variant="h5" component="h2" gutterBottom>
                  Submit Task
                </Typography>
                <TaskForm />
              </Paper>

              <TaskList sx={{ width: "50%" }} />
            </Box>

            <Box>
              <SystemStatus />
            </Box>
          </Box>

          <Box className="mt-[12px] text-center">
            <Typography variant="body2" color="text.secondary">
              Load Balancer Frontend • Built with React, TypeScript, Material-UI
              & Zustand
            </Typography>
          </Box>
        </Container>
      </Box>
    </ThemeProvider>
  );
}

export default App;
