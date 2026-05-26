import { defineStore } from "pinia";
import { apiRequest } from "@/api/http";

export const useCmdbApiStore = defineStore("cmdbApi", {
  state: () => ({
    assets: [],
    credentials: [],
    loading: false
  }),
  actions: {
    async fetchAssets() {
      const data = await apiRequest("/assets");
      this.assets = data.items || [];
      return this.assets;
    },
    async fetchCredentials() {
      const data = await apiRequest("/credentials");
      this.credentials = data.items || [];
      return this.credentials;
    },
    async createAsset(payload) {
      await apiRequest("/assets", { method: "POST", body: payload });
      await this.fetchAssets();
    },
    async updateAsset(id, payload) {
      await apiRequest(`/assets/${id}`, { method: "PUT", body: payload });
      await this.fetchAssets();
    },
    async deleteAsset(id) {
      await apiRequest(`/assets/${id}`, { method: "DELETE" });
      await this.fetchAssets();
    },
    async createCredential(payload) {
      await apiRequest("/credentials", { method: "POST", body: payload });
      await this.fetchCredentials();
    },
    async updateCredential(id, payload) {
      await apiRequest(`/credentials/${id}`, { method: "PUT", body: payload });
      await this.fetchCredentials();
    },
    async deleteCredential(id) {
      await apiRequest(`/credentials/${id}`, { method: "DELETE" });
      await this.fetchCredentials();
    },
    async sshTest(assetId, command) {
      return apiRequest("/ssh/test", {
        method: "POST",
        body: { assetId, command }
      });
    }
  }
});
