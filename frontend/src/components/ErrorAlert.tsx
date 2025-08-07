import React from "react";
import { Alert, AlertTitle, IconButton } from "@mui/material";
import { Close as CloseIcon } from "@mui/icons-material";
import { useTaskStore } from "../store/useTaskStore";

export const ErrorAlert: React.FC = () => {
  const { error, clearError } = useTaskStore();

  if (!error) return null;

  return (
    <Alert
      severity="error"
      className="mb-4"
      action={
        <IconButton
          aria-label="close"
          color="inherit"
          size="small"
          onClick={clearError}
        >
          <CloseIcon fontSize="inherit" />
        </IconButton>
      }
    >
      <AlertTitle>Error</AlertTitle>
      {error}
    </Alert>
  );
};
