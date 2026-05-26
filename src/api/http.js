const API_BASE = "/api";

export async function apiRequest(path, options = {}) {
  const token = localStorage.getItem("cmdb_token") || "";
  const method = options.method || "GET";
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {})
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    const error = new Error(data.error || "请求失败");
    error.status = response.status;
    error.payload = data;
    throw error;
  }

  return data;
}
