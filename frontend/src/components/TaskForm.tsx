import React, { useState } from "react";
import { TextField, Button, Box, CircularProgress } from "@mui/material";
import { Send as SendIcon } from "@mui/icons-material";
import { useTaskStore } from "../store/useTaskStore";

export const TaskForm: React.FC = () => {
  const [taskInput, setTaskInput] = useState("");
  const { submitTask, isLoading } = useTaskStore();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (taskInput.trim()) {
      await submitTask(taskInput.trim());
      setTaskInput("");
    }
  };

  return (
    <Box
      component="form"
      onSubmit={handleSubmit}
      className="mb-8 h-[fit-content] flex flex-col gap-100"
    >
      <Box className="flex" sx={{ gap: "20px" }}>
        <TextField
          fullWidth
          value={taskInput}
          onChange={(e) => setTaskInput(e.target.value)}
          placeholder="Enter a string to encrypt"
          disabled={isLoading}
          variant="outlined"
          size="medium"
          sx={{ flexGrow: 1 }}
        />
        <Button
          type="submit"
          disabled={isLoading || !taskInput.trim()}
          variant="contained"
          endIcon={
            isLoading ? (
              <CircularProgress size={20} color="inherit" />
            ) : (
              <SendIcon />
            )
          }
          sx={{ minWidth: "140px", height: "56px" }}
        >
          {isLoading ? "Submitting" : "Submit Task"}
        </Button>
      </Box>
    </Box>
  );
};
