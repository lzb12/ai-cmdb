<template>
  <section class="page-wrap">
    <article class="glass-panel toolbar fade-in-up">
      <input v-model.trim="keyword" placeholder="搜索凭据名称 / 用户名" />
      <button class="primary-btn" @click="openCreate">新增凭据</button>
    </article>

    <article class="glass-panel table-panel fade-in-up">
      <header>
        <h3>凭据管理</h3>
        <small>{{ filteredItems.length }} 条</small>
      </header>

      <div v-if="filteredItems.length" class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>认证方式</th>
              <th>用户名</th>
              <th>描述</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredItems" :key="item.id">
              <td><strong>{{ item.name }}</strong></td>
              <td>{{ item.authType === "password" ? "账号密码" : "SSH 私钥" }}</td>
              <td>{{ item.username }}</td>
              <td>{{ item.description || "--" }}</td>
              <td>{{ formatTime(item.updatedAt) }}</td>
              <td class="actions">
                <button class="ghost-btn" @click="openEdit(item)">编辑</button>
                <button class="ghost-btn danger" @click="remove(item)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty">暂无凭据数据</div>
    </article>

    <div v-if="formOpen" class="modal-mask" @click.self="formOpen = false">
      <div class="modal-panel slide-in-right">
        <header>
          <h3>{{ editingId ? "编辑凭据" : "新增凭据" }}</h3>
          <button class="ghost-btn" @click="formOpen = false">关闭</button>
        </header>

        <form @submit.prevent="saveCredential">
          <label>
            <span>凭据名称</span>
            <input v-model.trim="form.name" required />
          </label>
          <label>
            <span>认证方式</span>
            <select v-model="form.authType" :disabled="Boolean(editingId)">
              <option value="password">账号密码</option>
              <option value="key">SSH 私钥</option>
            </select>
          </label>
          <label>
            <span>登录用户名</span>
            <input v-model.trim="form.username" required />
          </label>
          <label>
            <span>描述</span>
            <input v-model.trim="form.description" />
          </label>

          <label v-if="form.authType === 'password'">
            <span>{{ editingId ? "密码（留空表示不修改）" : "密码" }}</span>
            <input v-model="form.password" type="password" :required="!editingId" />
          </label>

          <label v-if="form.authType === 'key'">
            <span>{{ editingId ? "私钥（留空表示不修改）" : "私钥" }}</span>
            <textarea v-model="form.privateKey" rows="6" :required="!editingId" />
          </label>

          <label v-if="form.authType === 'key'">
            <span>私钥口令（可选）</span>
            <input v-model="form.passphrase" type="password" />
          </label>

          <p v-if="errorText" class="error">{{ errorText }}</p>
          <footer>
            <button class="ghost-btn" type="button" @click="formOpen = false">取消</button>
            <button class="primary-btn" type="submit">保存</button>
          </footer>
        </form>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useCmdbApiStore } from "@/stores/cmdbApi";

const cmdbStore = useCmdbApiStore();

const keyword = ref("");
const formOpen = ref(false);
const editingId = ref(null);
const errorText = ref("");

const form = reactive({
  name: "",
  authType: "password",
  username: "",
  description: "",
  password: "",
  privateKey: "",
  passphrase: ""
});

const filteredItems = computed(() => {
  const key = keyword.value.toLowerCase();
  return cmdbStore.credentials.filter((item) => {
    if (!key) return true;
    return [item.name, item.username, item.description]
      .filter(Boolean)
      .join(" ")
      .toLowerCase()
      .includes(key);
  });
});

onMounted(async () => {
  await cmdbStore.fetchCredentials();
});

function openCreate() {
  editingId.value = null;
  form.name = "";
  form.authType = "password";
  form.username = "";
  form.description = "";
  form.password = "";
  form.privateKey = "";
  form.passphrase = "";
  errorText.value = "";
  formOpen.value = true;
}

function openEdit(item) {
  editingId.value = item.id;
  form.name = item.name || "";
  form.authType = item.authType || "password";
  form.username = item.username || "";
  form.description = item.description || "";
  form.password = "";
  form.privateKey = "";
  form.passphrase = "";
  errorText.value = "";
  formOpen.value = true;
}

async function saveCredential() {
  const payload = {
    name: form.name,
    authType: form.authType,
    username: form.username,
    description: form.description,
    password: form.password,
    privateKey: form.privateKey,
    passphrase: form.passphrase
  };

  try {
    if (editingId.value) {
      await cmdbStore.updateCredential(editingId.value, payload);
    } else {
      await cmdbStore.createCredential(payload);
    }
    formOpen.value = false;
  } catch (error) {
    errorText.value = error?.message || "保存失败";
  }
}

async function remove(item) {
  if (!window.confirm(`确认删除凭据 ${item.name} ?`)) return;
  try {
    await cmdbStore.deleteCredential(item.id);
  } catch (error) {
    window.alert(error?.message || "删除失败");
  }
}

function formatTime(value) {
  if (!value) return "--";
  const date = new Date(value);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(
    date.getDate()
  ).padStart(2, "0")} ${String(date.getHours()).padStart(2, "0")}:${String(
    date.getMinutes()
  ).padStart(2, "0")}`;
}
</script>

<style scoped>
.page-wrap {
  display: grid;
  gap: 12px;
}

.toolbar {
  padding: 12px;
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.toolbar input {
  flex: 1;
}

.table-panel {
  padding: 12px;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

header small {
  color: var(--text-soft);
}

.table-wrap {
  overflow: auto;
}

table {
  width: 100%;
  min-width: 820px;
  border-collapse: collapse;
}

th,
td {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 1px solid var(--line-soft);
}

.actions {
  display: flex;
  gap: 6px;
}

.danger {
  color: #99282f;
}

.empty {
  border: 1px dashed #afd0e5;
  border-radius: 10px;
  padding: 24px;
  text-align: center;
  color: var(--text-soft);
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(8, 33, 52, 0.44);
  z-index: 40;
  display: flex;
  justify-content: flex-end;
}

.modal-panel {
  width: min(640px, 100%);
  background: #f7fbff;
  padding: 16px;
  overflow: auto;
}

form {
  display: grid;
  gap: 10px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

span {
  font-size: 0.83rem;
  color: var(--text-soft);
}

textarea {
  border: 1px solid #b8d2e3;
  border-radius: 8px;
  padding: 10px 12px;
  font: inherit;
  resize: vertical;
}

textarea:focus {
  outline: none;
  border-color: #278bcb;
  box-shadow: 0 0 0 3px rgba(39, 139, 203, 0.2);
}

.error {
  color: #bc2b33;
}

footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 760px) {
  .toolbar {
    flex-direction: column;
  }
}
</style>
