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

// TRINI-specific interfaces
export interface GCSnapshot {
  timestamp: string;
  young_gen_used: number;
  old_gen_used: number;
  young_gen_max: number;
  old_gen_max: number;
  total_mem_used: number;
  total_mem_max: number;
  gc_count: number;
  last_magc_time: string;
  magc_duration_ms: number;
  is_collecting_gc: boolean;
}

export interface MaGCForecast {
  predicted_time: string;
  confidence: number;
  young_gen_threshold: number;
  time_to_magc_ms: number;
  forecast_created_at: string;
  is_predicted_within_threshold: boolean;
}

export interface ProgramFamily {
  id: string;
  name: string;
  description: string;
  magc_threshold_ms: number;
  forecast_window_size: number;
}

export interface LoadBalancingPolicy {
  algorithm: string;
  gc_aware: boolean;
  magc_threshold_ms: number;
  history_window_size: number;
}

export interface TRINIServerDetails {
  server_id: number;
  current_family: ProgramFamily | null;
  gc_history_count: number;
  last_magc_forecast: MaGCForecast | null;
  young_gen_used: number;
  old_gen_used: number;
  young_gen_max: number;
  old_gen_max: number;
  gc_count: number;
  weights: number;
}

export interface TRINIStatus {
  active: boolean;
  monitor_interval: string;
  analysis_interval: string;
  program_families: number;
  current_policy: LoadBalancingPolicy;
  servers: TRINIServerDetails[];
}

export interface ProgramFamiliesResponse {
  default_family: string;
  families: Record<string, ProgramFamily>;
}

export interface GCHistoryResponse {
  server_id: number;
  history_count: number;
  returned_count: number;
  gc_history: GCSnapshot[];
}

interface TaskStore {
  // State
  tasks: Task[];
  systemStatus: SystemStatus | null;
  isLoading: boolean;
  error: string | null;

  // TRINI State
  triniStatus: TRINIStatus | null;
  programFamilies: ProgramFamiliesResponse | null;
  gcHistories: Record<number, GCHistoryResponse>;
  isTRINILoading: boolean;
  triniError: string | null;

  // Actions
  submitTask: (taskInput: string) => Promise<void>;
  getSystemStatus: () => Promise<void>;
  pingServer: (serverId: number) => Promise<ServerStatus | null>;
  clearError: () => void;
  clearTasks: () => void;

  // TRINI Actions
  getTRINIStatus: () => Promise<void>;
  getProgramFamilies: () => Promise<void>;
  getGCHistory: (serverId: number, limit?: number) => Promise<void>;
  updateTRINIPolicy: (policy: LoadBalancingPolicy) => Promise<void>;
  toggleTRINI: (active: boolean) => Promise<void>;
  clearTRINIError: () => void;
}

const API_BASE_URL = "/api/v1";

export const useTaskStore = create<TaskStore>((set) => ({
  // Initial state
  tasks: [],
  systemStatus: null,
  isLoading: false,
  error: null,

  // TRINI Initial state
  triniStatus: null,
  programFamilies: null,
  gcHistories: {},
  isTRINILoading: false,
  triniError: null,

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

  // TRINI Actions
  getTRINIStatus: async () => {
    set({ isTRINILoading: true, triniError: null });

    try {
      const response = await fetch(`${API_BASE_URL}/trini/status`);

      if (!response.ok) {
        throw new Error("Failed to fetch TRINI status");
      }

      const data: TRINIStatus = await response.json();

      set({
        triniStatus: data,
        isTRINILoading: false,
      });
    } catch (error) {
      set({
        triniError:
          error instanceof Error ? error.message : "Unknown error occurred",
        isTRINILoading: false,
      });
    }
  },

  getProgramFamilies: async () => {
    set({ isTRINILoading: true, triniError: null });

    try {
      const response = await fetch(`${API_BASE_URL}/trini/families`);

      if (!response.ok) {
        throw new Error("Failed to fetch program families");
      }

      const data: ProgramFamiliesResponse = await response.json();

      set({
        programFamilies: data,
        isTRINILoading: false,
      });
    } catch (error) {
      set({
        triniError:
          error instanceof Error ? error.message : "Unknown error occurred",
        isTRINILoading: false,
      });
    }
  },

  getGCHistory: async (serverId: number, limit = 50) => {
    set({ isTRINILoading: true, triniError: null });

    try {
      const response = await fetch(
        `${API_BASE_URL}/server/${serverId}/gc-history?limit=${limit}`
      );

      if (!response.ok) {
        throw new Error("Failed to fetch GC history");
      }

      const data: GCHistoryResponse = await response.json();

      set((state) => ({
        gcHistories: {
          ...state.gcHistories,
          [serverId]: data,
        },
        isTRINILoading: false,
      }));
    } catch (error) {
      set({
        triniError:
          error instanceof Error ? error.message : "Unknown error occurred",
        isTRINILoading: false,
      });
    }
  },

  updateTRINIPolicy: async (policy: LoadBalancingPolicy) => {
    set({ isTRINILoading: true, triniError: null });

    try {
      const response = await fetch(`${API_BASE_URL}/trini/policy`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(policy),
      });

      if (!response.ok) {
        throw new Error("Failed to update TRINI policy");
      }

      // Refresh TRINI status after policy update
      const { getTRINIStatus } = useTaskStore.getState();
      await getTRINIStatus();

      set({ isTRINILoading: false });
    } catch (error) {
      set({
        triniError:
          error instanceof Error ? error.message : "Unknown error occurred",
        isTRINILoading: false,
      });
    }
  },

  toggleTRINI: async (active: boolean) => {
    set({ isTRINILoading: true, triniError: null });

    try {
      const response = await fetch(`${API_BASE_URL}/trini/toggle`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ active }),
      });

      if (!response.ok) {
        throw new Error("Failed to toggle TRINI");
      }

      // Refresh TRINI status after toggle
      const { getTRINIStatus } = useTaskStore.getState();
      await getTRINIStatus();

      set({ isTRINILoading: false });
    } catch (error) {
      set({
        triniError:
          error instanceof Error ? error.message : "Unknown error occurred",
        isTRINILoading: false,
      });
    }
  },

  clearTRINIError: () => set({ triniError: null }),
}));
