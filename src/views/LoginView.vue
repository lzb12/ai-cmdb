<template>
  <section class="login-page">
    <div class="login-card glass-panel fade-in-up">
      <h1>CMDB 登录</h1>
      <p>登录后进入主机资产管理页面。</p>

      <form @submit.prevent="submit">
        <label>
          <span>账号</span>
          <input v-model.trim="form.username" placeholder="请输入账号" />
        </label>
        <label>
          <span>密码</span>
          <input v-model="form.password" type="password" placeholder="请输入密码" />
        </label>
        <p v-if="errorText" class="error-text">{{ errorText }}</p>
        <button class="primary-btn" type="submit" :disabled="loading">
          {{ loading ? "登录中..." : "登录" }}
        </button>
      </form>

      <small>默认账号：admin / Admin@123456</small>
    </div>
  </section>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const router = useRouter();
const authStore = useAuthStore();

const form = reactive({
  username: "admin",
  password: "Admin@123456"
});
const loading = ref(false);
const errorText = ref("");

async function submit() {
  if (!form.username || !form.password) {
    errorText.value = "请输入账号和密码";
    return;
  }

  loading.value = true;
  errorText.value = "";
  try {
    await authStore.login(form.username, form.password);
    router.push("/assets");
  } catch (error) {
    errorText.value = error?.message || "登录失败";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background: radial-gradient(circle at 18% 12%, #d5ecfe 0, #ebf5ff 44%, #f8fcff 100%);
  padding: 16px;
}

.login-card {
  width: min(440px, 100%);
  padding: 24px;
  border-radius: 14px;
}

h1 {
  font-size: 1.5rem;
}

p {
  color: var(--text-soft);
  margin-top: 6px;
}

form {
  margin-top: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

span {
  font-size: 0.84rem;
  color: var(--text-soft);
}

button {
  margin-top: 4px;
}

.error-text {
  color: #be2e35;
}

small {
  margin-top: 10px;
  display: block;
  color: var(--text-soft);
}
</style>
