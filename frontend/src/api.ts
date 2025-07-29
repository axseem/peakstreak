export const api = {
  async post(path: string, body: any, token: string | null = null) {
    const headers: HeadersInit = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Request failed with status ${res.status}`);
    }
    return res.status === 204 ? null : await res.json();
  },
  async get(path: string, token: string | null) {
    const headers: HeadersInit = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, { headers });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "An unknown error occurred");
    return data;
  },
  async put(path: string, body: any, token: string | null) {
    const headers: HeadersInit = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, {
      method: "PUT",
      headers,
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Request failed with status ${res.status}`);
    }
    return res.status === 204 ? null : await res.json();
  },
  async delete(path: string, token: string | null) {
    const headers: HeadersInit = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, { method: "DELETE", headers });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Request failed with status ${res.status}`);
    }
    return res.status === 204 ? null : await res.json();
  },
  async upload(path: string, formData: FormData, token: string | null) {
    const headers: HeadersInit = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, { method: "POST", headers, body: formData });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Request failed with status ${res.status}`);
    }
    return res.json();
  },
};
