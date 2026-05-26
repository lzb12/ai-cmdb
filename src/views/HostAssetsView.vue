<template>
  <section class="page-wrap">
    <article class="glass-panel toolbar fade-in-up">
      <input v-model.trim="keyword" placeholder="搜索主机名 / 地址 / 负责人" />
      <button class="primary-btn" @click="openCreate">新增主机</button>
    </article>

    <article class="glass-panel table-panel fade-in-up">
      <header>
        <h3>主机资产</h3>
        <small>{{ filteredAssets.length }} 台</small>
      </header>

      <div v-if="filteredAssets.length" class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>主机名</th>
              <th>地址</th>
              <th>端口</th>
              <th>环境</th>
              <th>负责人</th>
              <th>凭据</th>
              <th>标签</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredAssets" :key="item.id">
              <td>
                <strong>{{ item.hostname }}</strong>
                <p>{{ item.os || "--" }}</p>
              </td>
              <td>{{ item.address }}</td>
              <td>{{ item.port }}</td>
              <td>{{ envText(item.environment) }}</td>
              <td>{{ item.owner }}</td>
              <td>{{ item.credential?.name || "未绑定" }}</td>
              <td>
                <span class="tag" v-for="tag in item.tags" :key="tag">{{ tag }}</span>
              </td>
              <td class="actions">
                <button class="ghost-btn" @click="openEdit(item)">编辑</button>
                <button class="ghost-btn" @click="runSSH(item)">SSH 测试</button>
                <button class="ghost-btn danger" @click="remove(item)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty">暂无资产数据</div>
    </article>

    <div v-if="formOpen" class="modal-mask" @click.self="formOpen = false">
      <div class="modal-panel slide-in-right">
        <header>
          <h3>{{ editingId ? "编辑主机" : "新增主机" }}</h3>
          <button class="ghost-btn" @click="formOpen = false">关闭</button>
        </header>

        <form @submit.prevent="saveAsset">
          <div class="form-grid">
            <label>
              <span>主机名</span>
              <input v-model.trim="form.hostname" required />
            </label>
            <label>
              <span>地址/IP</span>
              <input v-model.trim="form.address" required />
            </label>
            <label>
              <span>端口</span>
              <input v-model.number="form.port" type="number" min="1" />
            </label>
            <label>
              <span>操作系统</span>
              <input v-model.trim="form.os" />
            </label>
            <label>
              <span>环境</span>
              <select v-model="form.environment">
                <option value="prod">生产</option>
                <option value="staging">预发</option>
                <option value="test">测试</option>
              </select>
            </label>
            <label>
              <span>负责人</span>
              <input v-model.trim="form.owner" />
            </label>
            <label>
              <span>绑定凭据</span>
              <select v-model="credentialIdText">
                <option value="">不绑定</option>
                <option v-for="item in cmdbStore.credentials" :key="item.id" :value="String(item.id)">
                  {{ item.name }} ({{ item.username }})
                </option>
              </select>
            </label>
            <label>
              <span>标签（逗号分隔）</span>
              <input v-model="tagsText" placeholder="db,prod,critical" />
            </label>
          </div>

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
const tagsText = ref("");
const credentialIdText = ref("");

const form = reactive({
  hostname: "",
  address: "",
  port: 22,
  os: "",
  environment: "prod",
  owner: ""
});

const filteredAssets = computed(() => {
  const key = keyword.value.toLowerCase();
  return cmdbStore.assets.filter((item) => {
    if (!key) return true;
    return [item.hostname, item.address, item.owner, item.environment]
      .filter(Boolean)
      .join(" ")
      .toLowerCase()
      .includes(key);
  });
});

onMounted(async () => {
  await refresh();
});

async function refresh() {
  await Promise.all([cmdbStore.fetchAssets(), cmdbStore.fetchCredentials()]);
}

function openCreate() {
  editingId.value = null;
  form.hostname = "";
  form.address = "";
  form.port = 22;
  form.os = "";
  form.environment = "prod";
  form.owner = "";
  tagsText.value = "";
  credentialIdText.value = "";
  errorText.value = "";
  formOpen.value = true;
}

function openEdit(item) {
  editingId.value = item.id;
  form.hostname = item.hostname || "";
  form.address = item.address || "";
  form.port = Number(item.port) || 22;
  form.os = item.os || "";
  form.environment = item.environment || "prod";
  form.owner = item.owner || "";
  tagsText.value = Array.isArray(item.tags) ? item.tags.join(",") : "";
  credentialIdText.value = item.credentialId ? String(item.credentialId) : "";
  errorText.value = "";
  formOpen.value = true;
}

async function saveAsset() {
  const payload = {
    hostname: form.hostname,
    address: form.address,
    port: Number(form.port) || 22,
    os: form.os,
    environment: form.environment,
    owner: form.owner,
    credentialId: credentialIdText.value ? Number(credentialIdText.value) : null,
    tags: tagsText.value
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean)
  };

  try {
    if (editingId.value) {
      await cmdbStore.updateAsset(editingId.value, payload);
    } else {
      await cmdbStore.createAsset(payload);
    }
    formOpen.value = false;
  } catch (error) {
    errorText.value = error?.message || "保存失败";
  }
}

async function remove(item) {
  if (!window.confirm(`确认删除主机 ${item.hostname} ?`)) return;
  try {
    await cmdbStore.deleteAsset(item.id);
  } catch (error) {
    window.alert(error?.message || "删除失败");
  }
}

async function runSSH(item) {
  const command = window.prompt("输入要执行的命令（留空默认 hostname && uptime）", "") || "";
  try {
    const result = await cmdbStore.sshTest(item.id, command);
    window.alert(`执行成功:\n\n${result.output}`);
  } catch (error) {
    window.alert(`SSH 执行失败: ${error?.message || "未知错误"}`);
  }
}

function envText(env) {
  const map = { prod: "生产", staging: "预发", test: "测试" };
  return map[env] || env;
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
  min-width: 980px;
  border-collapse: collapse;
}

th,
td {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 1px solid var(--line-soft);
}

p {
  margin-top: 4px;
  font-size: 0.82rem;
  color: var(--text-soft);
}

.actions {
  display: flex;
  gap: 6px;
}

.tag {
  display: inline-block;
  margin: 2px 6px 2px 0;
  padding: 4px 8px;
  border-radius: 999px;
  background: #e7f2fa;
  color: #24587f;
  font-size: 0.75rem;
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
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

.error {
  margin-top: 10px;
  color: #bc2b33;
}

footer {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 760px) {
  .toolbar {
    flex-direction: column;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
