import { create } from "zustand";

export interface Task {
  id: string;
  input: string;
  output: string;
  status: "completed" | "rejected" | "pending";
  created_at: string;
}

export interface TaskResponse {
  status: string;
  message: string;
  task_id?: string;
  output?: string;
}

export interface ServerStatus {
  server_id: number;
  status: string;
  is_available: boolean;
  is_collecting_gc: boolean;
  mem_used: string;
  tasks_processed: number;
  task_ids: string[];
  memory_usage: string;
}

export interface SystemStatus {
  total_servers: number;
  available_servers: number;
  servers: ServerStatus[];
}

interface TaskStore {
  // State
  tasks: Task[];
  systemStatus: SystemStatus | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  submitTask: (taskInput: string) => Promise<void>;
  getSystemStatus: () => Promise<void>;
  pingServer: (serverId: number) => Promise<ServerStatus | null>;
  clearError: () => void;
  clearTasks: () => void;
}

const API_BASE_URL = "/api/v1";

export const useTaskStore = create<TaskStore>((set) => ({
  // Initial state
  tasks: [],
  systemStatus: null,
  isLoading: false,
  error: null,

  // Actions
  submitTask: async (taskInput: string) => {
    set({ isLoading: true, error: null });

    try {
      const response = await fetch(`${API_BASE_URL}/task`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ task: taskInput }),
      });

      const data: TaskResponse = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to submit task");
      }

      // Create a task object from the response
      const newTask: Task = {
        id: data.task_id || `task-${Date.now()}`,
        input: taskInput,
        output: data.output || "",
        status: data.status as Task["status"],
        created_at: new Date().toISOString(),
      };

      set((state) => ({
        tasks: [newTask, ...state.tasks],
        isLoading: false,
      }));
    } catch (error) {
      set({
        error:
          error instanceof Error ? error.message : "Unknown error occurred",
        isLoading: false,
      });
    }
  },

  getSystemStatus: async () => {
    set({ error: null });

    try {
      const response = await fetch(`${API_BASE_URL}/status`);

      if (!response.ok) {
        throw new Error("Failed to fetch system status");
      }

      const data: SystemStatus = await response.json();

      set({
        systemStatus: data,
        isLoading: false,
      });
    } catch (error) {
      set({
        error:
          error instanceof Error ? error.message : "Unknown error occurred",
        isLoading: false,
      });
    }
  },

  pingServer: async (serverId: number): Promise<ServerStatus | null> => {
    try {
      const response = await fetch(`${API_BASE_URL}/server/${serverId}/ping`);

      if (!response.ok) {
        throw new Error("Failed to ping server");
      }

      const data: ServerStatus = await response.json();
      return data;
    } catch (error) {
      set({
        error:
          error instanceof Error ? error.message : "Unknown error occurred",
      });
      return null;
    }
  },

  clearError: () => set({ error: null }),

  clearTasks: () => set({ tasks: [] }),
}));
