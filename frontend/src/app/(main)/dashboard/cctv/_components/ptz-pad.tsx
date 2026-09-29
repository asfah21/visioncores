"use client";

import { useEffect, useState } from "react";
import { ChevronDown, ChevronLeft, ChevronRight, ChevronUp, Home, Minus, Plus, Square } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cctvApi, type Preset } from "@/lib/cctv-api";
import { cn } from "@/lib/utils";

const SPEEDS = [
  { key: "slow", label: "Slow", scale: 0.35 },
  { key: "normal", label: "Normal", scale: 0.65 },
  { key: "fast", label: "Fast", scale: 1 },
] as const;

type SpeedKey = (typeof SPEEDS)[number]["key"];

// Remote-control style PTZ pad: preset grid, D-pad with home, zoom rocker,
// speed selector. Inspired by classic PTZ remote layouts.
export function PtzPad({ cameraId }: { cameraId: string }) {
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState("");
  const [speed, setSpeed] = useState<SpeedKey>("normal");
  const [presets, setPresets] = useState<Preset[] | null>(null);
  const [assigning, setAssigning] = useState(false);

  async function loadPresets() {
    try {
      setPresets((await cctvApi.ptzPresets(cameraId)) ?? []);
    } catch {
      setPresets(null);
    }
  }

  // biome-ignore lint/correctness/useExhaustiveDependencies: load on camera change only
  useEffect(() => {
    loadPresets();
  }, [cameraId]);

  function fail(e: unknown) {
    const raw = e instanceof Error ? e.message : "PTZ command failed";
    // The device answered but refuses the move: no PTZ hardware/service.
    // Show a friendly note instead of a raw SOAP fault.
    setMsg(/ptz move http \d+|ActionNotSupported|not supported|no ptz/i.test(raw) ? "PTZ is not supported by this camera." : raw);
  }

  async function move(pan: number, tilt: number, zoom = 0, duration = 800) {
    const scale = SPEEDS.find((s) => s.key === speed)?.scale ?? 0.65;
    setBusy(true);
    setMsg("");
    try {
      await cctvApi.ptzMove(cameraId, pan * scale, tilt * scale, zoom * scale, duration);
    } catch (e) {
      fail(e);
    } finally {
      setBusy(false);
    }
  }

  async function stop() {
    setMsg("");
    try {
      await cctvApi.ptzStop(cameraId);
    } catch (e) {
      fail(e);
    }
  }

  async function home() {
    setBusy(true);
    setMsg("");
    try {
      await cctvApi.ptzHome(cameraId);
    } catch (e) {
      fail(e);
    } finally {
      setBusy(false);
    }
  }

  async function pressPreset(slot: number) {
    const p = presets?.[slot];
    setMsg("");
    if (assigning) {
      setBusy(true);
      try {
        await cctvApi.ptzSetPreset(cameraId, `Preset ${slot + 1}`);
        setAssigning(false);
        await loadPresets();
      } catch (e) {
        fail(e);
      } finally {
        setBusy(false);
      }
      return;
    }
    if (!p) return;
    setBusy(true);
    try {
      await cctvApi.ptzGotoPreset(cameraId, p.token);
    } catch (e) {
      fail(e);
    } finally {
      setBusy(false);
    }
  }

  const showGrid = (presets !== null && presets.length > 0) || assigning;
  const dirBtn = "h-12 w-12 rounded-full [&_svg]:size-5";
  const rockerBtn = "h-10 w-16 [&_svg]:size-5";

  return (
    <div className="flex flex-col items-center gap-3 rounded-lg bg-zinc-950 p-4 text-zinc-100">
      {showGrid ? (
        <div className="flex w-full flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] tracking-wider text-zinc-400 uppercase">Presets</span>
            <Button
              size="sm"
              variant={assigning ? "destructive" : "secondary"}
              className="h-7 text-xs"
              disabled={busy}
              onClick={() => setAssigning((a) => !a)}
            >
              {assigning ? "Cancel assign" : "Assign"}
            </Button>
          </div>
          {assigning ? <p className="text-[11px] text-amber-400">Assign mode: press a slot to save the current position.</p> : null}
          <div className="grid grid-cols-3 gap-1.5">
            {Array.from({ length: 9 }, (_, i) => {
              const p = presets?.[i];
              return (
                <Button
                  key={`preset-slot-${i}`}
                  variant={p ? "secondary" : "outline"}
                  className={cn("h-9 font-mono text-sm", !p && !assigning && "border-zinc-800 text-zinc-600", assigning && "border-dashed border-amber-500/60")}
                  disabled={busy || (!p && !assigning)}
                  title={p?.name ?? `Empty slot ${i + 1}`}
                  onClick={() => pressPreset(i)}
                >
                  {i + 1}
                </Button>
              );
            })}
          </div>
        </div>
      ) : null}

      <div className="flex w-full items-center justify-around gap-2">
        <div className="grid grid-cols-3 items-center gap-1">
          <span />
          <Button className={dirBtn} variant="secondary" size="icon" disabled={busy} onClick={() => move(0, 1)} aria-label="Tilt up">
            <ChevronUp />
          </Button>
          <span />
          <Button className={dirBtn} variant="secondary" size="icon" disabled={busy} onClick={() => move(-1, 0)} aria-label="Pan left">
            <ChevronLeft />
          </Button>
          <Button className={dirBtn} variant="default" size="icon" disabled={busy} onClick={home} aria-label="Go to home position">
            <Home />
          </Button>
          <Button className={dirBtn} variant="secondary" size="icon" disabled={busy} onClick={() => move(1, 0)} aria-label="Pan right">
            <ChevronRight />
          </Button>
          <span />
          <Button className={dirBtn} variant="secondary" size="icon" disabled={busy} onClick={() => move(0, -1)} aria-label="Tilt down">
            <ChevronDown />
          </Button>
          <span />
        </div>
        <div className="flex flex-col items-center gap-1">
          <span className="text-[11px] tracking-wider text-zinc-400 uppercase">Zoom</span>
          <Button className={cn(rockerBtn, "rounded-t-xl rounded-b-sm")} variant="secondary" disabled={busy} onClick={() => move(0, 0, 1)} aria-label="Zoom in">
            <Plus />
          </Button>
          <Button className={cn(rockerBtn, "rounded-t-sm rounded-b-xl")} variant="secondary" disabled={busy} onClick={() => move(0, 0, -1)} aria-label="Zoom out">
            <Minus />
          </Button>
        </div>
      </div>

      <div className="flex w-full items-center justify-between gap-2">
        <div className="flex gap-1">
          {SPEEDS.map((s) => (
            <Button
              key={s.key}
              size="sm"
              variant={speed === s.key ? "default" : "outline"}
              className={cn("h-7 px-2 text-xs", speed !== s.key && "border-zinc-700 text-zinc-300")}
              disabled={busy}
              onClick={() => setSpeed(s.key)}
            >
              {s.label}
            </Button>
          ))}
        </div>
        <Button variant="destructive" size="icon" className="h-9 w-9" onClick={stop} aria-label="Stop movement">
          <Square />
        </Button>
      </div>
      {msg ? <p className="w-full text-xs text-red-400">{msg}</p> : null}
    </div>
  );
}
