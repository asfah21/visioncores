"use client";

import { useState } from "react";
import { ChevronDown, ChevronLeft, ChevronRight, ChevronUp, Home, Square } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cctvApi } from "@/lib/cctv-api";

export function PtzPad({ cameraId }: { cameraId: string }) {
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState("");

  async function move(pan: number, tilt: number, zoom = 0, duration = 800) {
    setBusy(true);
    setMsg("");
    try {
      await cctvApi.ptzMove(cameraId, pan, tilt, zoom, duration);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "PTZ failed");
    } finally {
      setBusy(false);
    }
  }

  async function stop() {
    try {
      await cctvApi.ptzStop(cameraId);
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Stop failed");
    }
  }

  const btn = "h-11 w-11";
  return (
    <div className="flex flex-col items-center gap-2">
      <div className="grid grid-cols-3 items-center gap-1">
        <span />
        <Button className={btn} variant="outline" size="icon" disabled={busy} onClick={() => move(0, 1)} aria-label="Tilt up">
          <ChevronUp />
        </Button>
        <span />
        <Button className={btn} variant="outline" size="icon" disabled={busy} onClick={() => move(-1, 0)} aria-label="Pan left">
          <ChevronLeft />
        </Button>
        <Button className={btn} variant="destructive" size="icon" onClick={stop} aria-label="Stop">
          <Square />
        </Button>
        <Button className={btn} variant="outline" size="icon" disabled={busy} onClick={() => move(1, 0)} aria-label="Pan right">
          <ChevronRight />
        </Button>
        <span />
        <Button className={btn} variant="outline" size="icon" disabled={busy} onClick={() => move(0, -1)} aria-label="Tilt down">
          <ChevronDown />
        </Button>
        <span />
      </div>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" disabled={busy} onClick={() => move(0, 0, 1)}>
          Zoom [+]
        </Button>
        <Button variant="outline" size="sm" disabled={busy} onClick={() => move(0, 0, -1)}>
          Zoom [-]
        </Button>
        <Button variant="ghost" size="sm" disabled={busy} onClick={() => move(0, 0, 0, 300)}>
          <Home className="mr-1 size-4" /> Home
        </Button>
      </div>
      {msg ? <p className="text-destructive text-xs">{msg}</p> : null}
      <p className="text-muted-foreground text-[11px]">Pan/tilt/zoom ±1, auto-stop 800ms (maks 5s di backend).</p>
    </div>
  );
}
