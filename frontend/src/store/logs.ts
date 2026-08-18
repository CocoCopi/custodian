import { create } from "zustand";
import type { LogEntry } from "../api/types";

interface LogState {
  entries: LogEntry[];
  connected: boolean;
  append: (entry: LogEntry) => void;
  setConnected: (connected: boolean) => void;
  clear: () => void;
}

const MAX_ENTRIES = 2000;

export const useLogStore = create<LogState>((set) => ({
  entries: [],
  connected: false,
  append: (entry) =>
    set((state) => ({
      entries: [...state.entries, entry].slice(-MAX_ENTRIES),
    })),
  setConnected: (connected) => set({ connected }),
  clear: () => set({ entries: [] }),
}));
