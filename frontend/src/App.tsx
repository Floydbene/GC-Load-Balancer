import React, { useState } from "react";
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
  Button,
} from "@mui/material";
import { Psychology as PsychologyIcon } from "@mui/icons-material";
import { TaskFormWrapper } from "./components/TaskFormWrapper";
import { TaskForm } from "./components/TaskForm";
import { TaskList } from "./components/TaskList";
import { SystemStatus } from "./components/SystemStatus";
import { ErrorAlert } from "./components/ErrorAlert";
import { TRINIDashboard } from "./components/TRINIDashboard";

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
  const [showTRINI, setShowTRINI] = useState(false);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box className="min-h-screen bg-gray-50">
        <AppBar position="static" elevation={1}>
          <Toolbar
            sx={{
              justifyContent: "space-between",
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

            <Box sx={{ display: "flex", gap: 2 }}>
              <Button
                color="inherit"
                onClick={() => setShowTRINI(false)}
                variant={!showTRINI ? "outlined" : "text"}
                sx={{
                  borderColor: !showTRINI ? "white" : "transparent",
                  color: "white",
                }}
              >
                Task Manager
              </Button>
              <Button
                color="inherit"
                onClick={() => setShowTRINI(true)}
                variant={showTRINI ? "outlined" : "text"}
                sx={{
                  borderColor: showTRINI ? "white" : "transparent",
                  color: "white",
                }}
                startIcon={<PsychologyIcon />}
              >
                TRINI Dashboard
              </Button>
            </Box>
          </Toolbar>
        </AppBar>

        <Container maxWidth="xl" sx={{ py: 4 }}>
          <ErrorAlert />

          {!showTRINI ? (
            <Box className="grid grid-cols-1 lg:grid-cols-1 gap-8">
              <Box
                sx={{
                  display: "flex",
                  flexDirection: "column",
                  gap: 2,
                }}
              >
                <TaskFormWrapper>
                  <TaskForm />
                </TaskFormWrapper>

                <TaskList />
              </Box>

              <Box>
                <SystemStatus />
              </Box>
            </Box>
          ) : (
            <TRINIDashboard />
          )}

          <Box className="mt-[12px] text-center">
            <Typography variant="body2" color="text.secondary">
              Load Balancer Frontend • Built with React, TypeScript, Material-UI
              & Zustand •{" "}
              {showTRINI ? "TRINI GC-Aware Monitoring" : "Task Management"}
            </Typography>
          </Box>
        </Container>
      </Box>
    </ThemeProvider>
  );
}

export default App;
