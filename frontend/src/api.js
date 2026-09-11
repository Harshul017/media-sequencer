const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080';

async function request(path, options) {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...(options?.headers || {}) },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Request failed');
  return data;
}

export function getWindows() {
  return request('/windows');
}

export function addMedia(windowId, { type, url, duration_seconds }) {
  return request(`/windows/${windowId}/media`, {
    method: 'POST',
    body: JSON.stringify({ type, url, duration_seconds }),
  });
}

export function getSyncState() {
  return request('/sync-state');
}

export function triggerSync({ type, url, duration_seconds }) {
  return request('/sync', {
    method: 'POST',
    body: JSON.stringify({ type, url, duration_seconds }),
  });
}
