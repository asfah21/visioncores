"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { cctvApi } from "@/lib/cctv-api";

const empty = { id: "", name: "", description: "", host: "", rtsp_port: 8554, rtsp_url: "", username: "", password: "", onvif_port: 8000 };

export function AddCameraDialog({ open, onOpenChange, onAdded }: { open: boolean; onOpenChange: (v: boolean) => void; onAdded: () => void }) {
  const [form, setForm] = useState(empty);
  const [msg, setMsg] = useState("");
  const [saving, setSaving] = useState(false);

  function set(k: string, v: string | number) {
    setForm((f) => ({ ...f, [k]: v }));
  }

  function close(v: boolean) {
    if (!saving) {
      if (!v) {
        setForm(empty);
        setMsg("");
      }
      onOpenChange(v);
    }
  }

  async function submit() {
    setMsg("");
    if (!form.id.trim() || !form.name.trim()) {
      setMsg("Camera ID and Name are required.");
      return;
    }
    if (!form.rtsp_url.trim() && !form.host.trim()) {
      setMsg("Either RTSP URL or Host is required.");
      return;
    }
    setSaving(true);
    try {
      await cctvApi.createCamera({ ...form, id: form.id.trim() });
      try {
        await cctvApi.go2rtcSync();
      } catch {
        // non-fatal: camera is saved, stream registers on next sync/restart
      }
      setForm(empty);
      onOpenChange(false);
      onAdded();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Failed to add camera.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[560px]">
        <DialogHeader>
          <DialogTitle>Add Camera</DialogTitle>
          <DialogDescription>Register a new camera. The live stream becomes available after the stream config syncs.</DialogDescription>
        </DialogHeader>
        <div className="grid grid-cols-2 gap-3">
          <label htmlFor="cam-id" className="flex flex-col gap-1 text-xs">
            Camera ID *
            <Input id="cam-id" value={form.id} onChange={(e) => set("id", e.target.value)} placeholder="cam2" />
          </label>
          <label htmlFor="cam-name" className="flex flex-col gap-1 text-xs">
            Name *
            <Input id="cam-name" value={form.name} onChange={(e) => set("name", e.target.value)} placeholder="Cam 2" />
          </label>
          <label htmlFor="cam-desc" className="col-span-2 flex flex-col gap-1 text-xs">
            Description
            <Input id="cam-desc" value={form.description} onChange={(e) => set("description", e.target.value)} placeholder="Lobby entrance" />
          </label>
          <label htmlFor="cam-host" className="flex flex-col gap-1 text-xs">
            Host
            <Input id="cam-host" value={form.host} onChange={(e) => set("host", e.target.value)} placeholder="10.10.11.20" />
          </label>
          <label htmlFor="cam-rtsp-port" className="flex flex-col gap-1 text-xs">
            RTSP Port
            <Input id="cam-rtsp-port" type="number" value={form.rtsp_port} onChange={(e) => set("rtsp_port", Number(e.target.value))} />
          </label>
          <label htmlFor="cam-rtsp-url" className="col-span-2 flex flex-col gap-1 text-xs">
            RTSP URL (full, overrides host/port)
            <Input id="cam-rtsp-url" value={form.rtsp_url as string} onChange={(e) => set("rtsp_url", e.target.value)} placeholder="rtsp://user:pass@host:8554/Streaming/Channels/101" />
          </label>
          <label htmlFor="cam-user" className="flex flex-col gap-1 text-xs">
            Username
            <Input id="cam-user" value={form.username} onChange={(e) => set("username", e.target.value)} placeholder="admin" />
          </label>
          <label htmlFor="cam-pass" className="flex flex-col gap-1 text-xs">
            Password
            <Input id="cam-pass" type="password" value={form.password as string} onChange={(e) => set("password", e.target.value)} placeholder="••••••••" />
          </label>
          <label htmlFor="cam-onvif-port" className="flex flex-col gap-1 text-xs">
            ONVIF Port
            <Input id="cam-onvif-port" type="number" value={form.onvif_port} onChange={(e) => set("onvif_port", Number(e.target.value))} />
          </label>
        </div>
        {msg ? <p className="text-sm text-amber-500">{msg}</p> : null}
        <DialogFooter>
          <Button variant="outline" onClick={() => close(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={submit} disabled={saving}>
            {saving ? "Adding…" : "Add Camera"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
