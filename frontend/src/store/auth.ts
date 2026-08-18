import { create } from "zustand";

const TOKEN_KEY = "custodian_token";

interface AuthState {
  token: string | null;
  setToken: (token: string) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem(TOKEN_KEY),
  setToken: (token) => {
    localStorage.setItem(TOKEN_KEY, token);
    set({ token });
  },
  clear: () => {
    localStorage.removeItem(TOKEN_KEY);
    set({ token: null });
  },
}));
