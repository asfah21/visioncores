"use client";

import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { cctvApi, type PublicCamera } from "@/lib/cctv-api";

const empty = { id: "", name: "", description: "", host: "", rtsp_port: 8554, rtsp_url: "", username: "", password: "", onvif_port: 8000 };

export default function CamerasPage() {
  const [list, setList] = useState<PublicCamera[]>([]);
  const [form, setForm] = useState(empty);
  const [editing, setEditing] = useState<string | null>(null);
  const [msg, setMsg] = useState("");

  async function load() {
    try {
      setList(await cctvApi.cameras());
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "gagal memuat");
    }
  }
  useEffect(() => {
    load();
  }, []);

  function set(k: string, v: string | number) {
    setForm((f) => ({ ...f, [k]: v }));
  }

  async function submit() {
    setMsg("");
    try {
      if (editing) {
        await cctvApi.updateCamera(editing, { ...form, id: editing });
      } else {
        await cctvApi.createCamera(form);
      }
      setForm(empty);
      setEditing(null);
      load();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "simpan gagal");
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="font-bold text-xl">CCTV — Kelola Kamera (DB source of truth)</h1>
      {msg ? <p className="text-sm text-amber-500">{msg}</p> : null}
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Daftar</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            {list.map((c) => (
              <div key={c.id} className="flex items-center justify-between rounded border p-2 text-sm">
                <div>
                  <p className="font-semibold">
                    {c.name} <span className="text-muted-foreground">({c.id})</span>
                  </p>
                  <p className="text-muted-foreground text-xs">
                    {c.host}:{c.rtsp_port} · ONVIF:{c.onvif_port} · {c.auto_record ? "auto-rec" : "manual"}
                  </p>
                </div>
                <div className="flex gap-1">
                  <Button size="sm" variant="outline" onClick={() => cctvApi.testCamera(c.id).then((r) => setMsg(`${c.id}: RTSP ${r.rtsp.reachable ? "OK" : "gagal"} / ONVIF ${r.onvif.reachable ? "OK" : "gagal"}`)).catch((e) => setMsg(e.message))}>
                    Test
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => {
                      setEditing(c.id);
                      setForm({ id: c.id, name: c.name, description: c.description, host: c.host, rtsp_port: c.rtsp_port, rtsp_url: "", username: c.username, password: "", onvif_port: c.onvif_port });
                    }}
                  >
                    Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={async () => {
                      if (confirm(`Hapus ${c.id}?`)) {
                        await cctvApi.deleteCamera(c.id);
                        load();
                      }
                    }}
                  >
                    Hapus
                  </Button>
                </div>
              </div>
            ))}
            {list.length === 0 ? <p className="text-muted-foreground text-sm">Belum ada kamera.</p> : null}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{editing ? `Edit ${editing}` : "Tambah kamera"}</CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-2">
            <label className="flex flex-col gap-1 text-xs">ID<Input value={form.id} disabled={!!editing} onChange={(e) => set("id", e.target.value)} placeholder="cam1" /></label>
            <label className="flex flex-col gap-1 text-xs">Nama<Input value={form.name} onChange={(e) => set("name", e.target.value)} placeholder="Cam 1" /></label>
            <label className="col-span-2 flex flex-col gap-1 text-xs">Deskripsi<Input value={form.description} onChange={(e) => set("description", e.target.value)} /></label>
            <label className="flex flex-col gap-1 text-xs">Host<Input value={form.host} onChange={(e) => set("host", e.target.value)} placeholder="10.10.11.20" /></label>
            <label className="flex flex-col gap-1 text-xs">RTSP port<Input type="number" value={form.rtsp_port} onChange={(e) => set("rtsp_port", Number(e.target.value))} /></label>
            <label className="col-span-2 flex flex-col gap-1 text-xs">RTSP URL (penuh, password tidak ditampilkan kembali)<Input value={form.rtsp_url as string} onChange={(e) => set("rtsp_url", e.target.value)} placeholder="rtsp://user:pass@host:8554/Streaming/Channels/101" /></label>
            <label className="flex flex-col gap-1 text-xs">Username<Input value={form.username} onChange={(e) => set("username", e.target.value)} /></label>
            <label className="flex flex-col gap-1 text-xs">Password (baru)<Input type="password" value={form.password as string} onChange={(e) => set("password", e.target.value)} /></label>
            <label className="flex flex-col gap-1 text-xs">ONVIF port<Input type="number" value={form.onvif_port} onChange={(e) => set("onvif_port", Number(e.target.value))} /></label>
            <div className="col-span-2 flex gap-2">
              <Button onClick={submit}>{editing ? "Simpan" : "Tambah"}</Button>
              {editing ? <Button variant="ghost" onClick={() => { setEditing(null); setForm(empty); }}>Batal</Button> : null}
            </div>
            <p className="col-span-2 text-[11px] text-muted-foreground">Password terenkripsi AES-GCM di Postgres, tidak pernah dikembalikan ke frontend. RTSP full hanya dipakai backend untuk ffmpeg/go2rtc.</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
