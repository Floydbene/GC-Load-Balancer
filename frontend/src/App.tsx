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
      main: "#007d9c",
    },
    secondary: {
      main: "#007d9c",
    },
  },
});

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box className="min-h-screen bg-gray-50">
        <AppBar position="static" elevation={1}>
          <Toolbar
            sx={{
              justifyContent: "center",
              minHeight: "64px",
              backgroundColor: "#007d9c",
            }}
          >
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <img
                src="https://go.dev/blog/go-brand/Go-Logo/SVG/Go-Logo_Fuchsia.svg"
                alt="Go Logo"
                style={{ height: 60 }}
              />
              <Typography
                variant="h5"
                component="div"
                sx={{ flexGrow: 1, my: "20px" }}
              >
                GC Load Balancer
              </Typography>
            </Box>
          </Toolbar>
        </AppBar>

        <Container maxWidth="lg" sx={{ py: 4 }}>
          <ErrorAlert />

          <Box className="grid grid-cols-1 lg:grid-cols-1 gap-8">
            <Box
              sx={{
                display: "flex",
                flexDirection: "column",
                gap: 2,
              }}
            >
              <Paper elevation={1} sx={{ p: 3, mb: 3, height: "fit-content" }}>
                <Typography variant="h5" component="h2" gutterBottom>
                  Submit Task
                </Typography>
                <TaskForm />
              </Paper>

              <TaskList />
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
