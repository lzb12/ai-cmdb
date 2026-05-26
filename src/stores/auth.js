import { defineStore } from "pinia";
import { apiRequest } from "@/api/http";

const TOKEN_KEY = "cmdb_token";
const USER_KEY = "cmdb_user";

function safeParseJSON(raw, fallback) {
  try {
    return raw ? JSON.parse(raw) : fallback;
  } catch {
    return fallback;
  }
}

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || "",
    user: safeParseJSON(localStorage.getItem(USER_KEY), null),
    ready: false
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token)
  },
  actions: {
    setAuth(token, user) {
      this.token = token || "";
      this.user = user || null;
      if (this.token) {
        localStorage.setItem(TOKEN_KEY, this.token);
      } else {
        localStorage.removeItem(TOKEN_KEY);
      }
      if (this.user) {
        localStorage.setItem(USER_KEY, JSON.stringify(this.user));
      } else {
        localStorage.removeItem(USER_KEY);
      }
    },
    clearAuth() {
      this.setAuth("", null);
    },
    async login(username, password) {
      const data = await apiRequest("/auth/login", {
        method: "POST",
        body: { username, password }
      });
      this.setAuth(data.token, data.user);
      return data;
    },
    async fetchMe() {
      if (!this.token) {
        this.ready = true;
        return null;
      }
      try {
        const data = await apiRequest("/auth/me");
        this.user = data.user || this.user;
        localStorage.setItem(USER_KEY, JSON.stringify(this.user));
        return data.user;
      } catch (error) {
        if (error?.status === 401) {
          this.clearAuth();
        }
        throw error;
      } finally {
        this.ready = true;
      }
    },
    async logout() {
      try {
        if (this.token) {
          await apiRequest("/auth/logout", { method: "POST" });
        }
      } finally {
        this.clearAuth();
      }
    }
  }
});
