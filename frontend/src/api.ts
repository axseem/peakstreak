export const api = {
  async post(path: string, body: any, token: string | null = null) {
    const headers: HeadersInit = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, { method: "POST", headers, body: JSON.stringify(body) });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "An unknown error occurred");
    return data;
  },
  async get(path: string, token: string | null) {
    const headers: HeadersInit = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    const res = await fetch(path, { headers });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "An unknown error occurred");
    return data;
  }
};
