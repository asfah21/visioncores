"use client";

export type PublicCamera = {
  id: string;
  name: string;
  description: string;
  host: string;
  rtsp_port: number;
  onvif_port: number;
  username: string;
  enabled: boolean;
  auto_record: boolean;
  created_at?: string;
  updated_at?: string;
};

export type RecordingItem = {
  id: number;
  camera_id: string;
  start_time: string;
  end_time?: string;
  duration: number;
  file_size: number;
  format: string;
  status: string;
};

export type Segment = { file: string; start: string; size: number };

export type Preset = { token: string; name: string };

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const r = await fetch(`/api/cctv/${path}`, {
    ...init,
    headers: { "content-type": "application/json", ...(init?.headers ?? {}) },
  });
  if (r.status === 204) return undefined as T;
  const j = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error((j as { error?: string }).error ?? `request failed ${r.status}`);
  return j as T;
}

export const cctvApi = {
  cameras: () => req<PublicCamera[]>("cameras"),
  camera: (id: string) => req<PublicCamera>(`cameras/${id}`),
  createCamera: (b: Record<string, unknown>) => req<PublicCamera>("cameras", { method: "POST", body: JSON.stringify(b) }),
  updateCamera: (id: string, b: Record<string, unknown>) =>
    req<PublicCamera>(`cameras/${id}`, { method: "PUT", body: JSON.stringify(b) }),
  deleteCamera: (id: string) => req<void>(`cameras/${id}`, { method: "DELETE" }),
  testCamera: (id: string) =>
    req<{ rtsp: { reachable: boolean }; onvif: { reachable: boolean } }>(`cameras/${id}/test-connection`, {
      method: "POST",
      body: "{}",
    }),
  streamUrl: (id: string) => req<{ webrtc: string; mse: string; hls: string }>(`cameras/${id}/stream-url`),
  info: (id: string) => req<{ device: { manufacturer: string; model: string; firmware: string }; profiles: { token: string; name: string }[]; capabilities: { ptz: boolean } }>(`cameras/${id}/info`),
  ptzMove: (id: string, pan: number, tilt: number, zoom: number, duration = 800) =>
    req<{ ok: boolean }>(`cameras/${id}/ptz/move`, { method: "POST", body: JSON.stringify({ pan, tilt, zoom, duration }) }),
  ptzStop: (id: string) => req<{ ok: boolean }>(`cameras/${id}/ptz/stop`, { method: "POST", body: "{}" }),
  ptzStatus: (id: string) => req<{ pan: number; tilt: number; zoom: number }>(`cameras/${id}/ptz/status`),
  ptzPresets: (id: string) => req<Preset[]>(`cameras/${id}/ptz/presets`),
  ptzGotoPreset: (id: string, preset: string) =>
    req<{ ok: boolean }>(`cameras/${id}/ptz/goto`, { method: "POST", body: JSON.stringify({ preset }) }),
  ptzSetPreset: (id: string, name: string) =>
    req<{ ok: boolean; preset: string }>(`cameras/${id}/ptz/preset`, { method: "POST", body: JSON.stringify({ name }) }),
  ptzHome: (id: string) => req<{ ok: boolean }>(`cameras/${id}/ptz/home`, { method: "POST", body: "{}" }),
  recordings: (id: string, q = "") =>
    req<{ recordings: RecordingItem[]; segments: Segment[]; recording: boolean }>(`cameras/${id}/recordings${q}`),
  timeline: (id: string, date: string) =>
    req<{ date: string; recordings: RecordingItem[]; segments: Segment[] }>(`cameras/${id}/recordings/timeline?date=${date}`),
  recording: (id: number) => req<RecordingItem>(`recordings/${id}`),
  events: (id: number) =>
    req<{ camera_id: string; person_id: number; action: string; at: string }[]>(`recordings/${id}/events`),
  startRec: (id: string) => req<{ ok: boolean }>(`cameras/${id}/recordings/start`, { method: "POST", body: "{}" }),
  stopRec: (id: string) => req<{ ok: boolean }>(`cameras/${id}/recordings/stop`, { method: "POST", body: "{}" }),
  go2rtcSync: () => req<{ ok: boolean; streams: number }>(`go2rtc/sync`, { method: "POST", body: "{}" }),
};

export function streamFileUrl(recId: number, file: string) {
  return `/api/cctv/recordings/${recId}/stream?file=${encodeURIComponent(file)}`;
}
