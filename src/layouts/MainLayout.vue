<template>
  <div class="app-shell">
    <aside class="sidebar glass-panel fade-in-up">
      <div class="brand">
        <div class="mark">CM</div>
        <div>
          <h1>Ops CMDB</h1>
          <p>资产与凭据中心</p>
        </div>
      </div>

      <nav class="nav-list">
        <RouterLink to="/assets" class="nav-item">主机资产</RouterLink>
        <RouterLink to="/credentials" class="nav-item">凭据管理</RouterLink>
      </nav>

      <button class="ghost-btn logout-btn" @click="doLogout">退出登录</button>
    </aside>

    <div class="main-wrap">
      <header class="topbar fade-in-up">
        <div>
          <small>当前模块</small>
          <h2>{{ route.meta.title || "CMDB" }}</h2>
        </div>
        <div class="user-chip">
          <span class="dot" />
          <strong>{{ username }}</strong>
        </div>
      </header>

      <main class="content-area">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const username = computed(() => authStore.user?.username || "未知用户");

async function doLogout() {
  await authStore.logout();
  router.push("/login");
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  background: radial-gradient(circle at 12% 8%, #d8efff 0, #edf6ff 38%, #f7fbff 100%);
}

.sidebar {
  padding: 20px 14px;
  border-right: 1px solid var(--line-soft);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mark {
  width: 42px;
  height: 42px;
  border-radius: 11px;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  background: linear-gradient(140deg, #0e456f, #2394d3);
}

.brand h1 {
  font-size: 1rem;
}

.brand p {
  color: var(--text-soft);
  margin-top: 4px;
  font-size: 0.82rem;
}

.nav-list {
  display: grid;
  gap: 8px;
}

.nav-item {
  border-radius: 10px;
  padding: 11px 12px;
  color: var(--text-primary);
  font-weight: 600;
  transition: all 0.2s;
}

.nav-item:hover {
  background: rgba(17, 102, 157, 0.1);
}

.nav-item.router-link-active {
  background: linear-gradient(130deg, #134970, #1f7eb6);
  color: #fff;
}

.logout-btn {
  margin-top: auto;
}

.main-wrap {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  padding: 16px 20px;
  border-bottom: 1px solid var(--line-soft);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.topbar small {
  color: var(--text-soft);
}

.topbar h2 {
  margin-top: 4px;
  font-size: 1.4rem;
}

.user-chip {
  border-radius: 999px;
  border: 1px solid var(--line-soft);
  background: rgba(255, 255, 255, 0.8);
  padding: 8px 12px;
  display: inline-flex;
  gap: 8px;
  align-items: center;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #18a760;
}

.content-area {
  padding: 16px 20px 26px;
}

@media (max-width: 900px) {
  .app-shell {
    grid-template-columns: 1fr;
  }

  .sidebar {
    border-right: 0;
    border-bottom: 1px solid var(--line-soft);
  }

  .content-area {
    padding: 14px;
  }
}
</style>
