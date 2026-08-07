"use client";

import { useEffect, useRef, useState } from "react";

import { Check, ChevronsUpDown, Circle, Maximize, Minimize, ShieldAlert } from "lucide-react";

import { AspectRatio } from "@/components/ui/aspect-ratio";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader } from "@/components/ui/card";
import { Command, CommandGroup, CommandItem, CommandList } from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

import { CAMERAS } from "./generated-cameras";

type Detection = {
  bbox: { x1: number; y1: number; x2: number; y2: number };
  name: string;
  person_id: number;
  frame_size: { w: number; h: number };
  camera_id?: string;
};

export function LiveCctvCard({
  className,
  selectedCam,
  onCamChange,
}: {
  className?: string;
  selectedCam: string;
  onCamChange: (val: string) => void;
}) {
  const [detections, setDetections] = useState<Detection[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [camStatus, setCamStatus] = useState<Record<string, boolean>>({});
  const [isFullscreen, setIsFullscreen] = useState(false);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const selectedCamRef = useRef(selectedCam);
  const camTimeoutsRef = useRef<Record<string, NodeJS.Timeout>>({});
  selectedCamRef.current = selectedCam;

  useEffect(() => {
    const saved = localStorage.getItem("visioncore:selectedCam");
    if (saved && saved !== selectedCam) onCamChange(saved);
  }, [onCamChange, selectedCam]);

  const handleCamChange = (val: string) => {
    onCamChange(val);
    setDetections([]);
    localStorage.setItem("visioncore:selectedCam", val);
  };

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
  }, []);

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen().catch((err) => {
        console.error("Fullscreen error:", err);
      });
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
  };

  // ===== WEBSOCKET =====
  useEffect(() => {
    const ws = new WebSocket("ws://10.10.11.5:3001/ws");

    ws.onopen = () => setIsConnected(true);
    ws.onclose = () => {
      setIsConnected(false);
      setCamStatus({});
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        // Ambil ID kamera dari root, atau dari deteksi pertama
        let camCode = data.camera_id;
        if (
          !camCode &&
          data.event === "person_detect_batch" &&
          Array.isArray(data.detections) &&
          data.detections.length > 0
        ) {
          camCode = data.detections[0].camera_id;
        }

        // --- MANAJEMEN STATUS ONLINE/OFFLINE PER KAMERA ---
        if (camCode) {
          setCamStatus((prev) => {
            if (!prev[camCode]) return { ...prev, [camCode]: true };
            return prev;
          });

          if (camTimeoutsRef.current[camCode]) {
            clearTimeout(camTimeoutsRef.current[camCode]);
          }

          camTimeoutsRef.current[camCode] = setTimeout(() => {
            setCamStatus((prev) => ({ ...prev, [camCode]: false }));
            if (selectedCamRef.current === camCode) {
              setDetections([]);
              window.dispatchEvent(new CustomEvent("ai:detections", { detail: [] }));
            }
          }, 1500); // 1.5 detik tidak ada update untuk kamera ini = kosongkan & offline
        }

        // --- MANAJEMEN DETEKSI (GAMBAR BOX) ---
        if (data.event === "person_detect_batch" && Array.isArray(data.detections)) {
          if (camCode === selectedCamRef.current || !camCode) {
            const filtered = data.detections.filter((d: any) => d.camera_id === selectedCamRef.current);
            setDetections(filtered);
            window.dispatchEvent(new CustomEvent("ai:detections", { detail: filtered }));
          }
        } else if (data.event === "person_detect") {
          if (data.camera_id === selectedCamRef.current) {
            const single = [data];
            setDetections(single);
            window.dispatchEvent(new CustomEvent("ai:detections", { detail: single }));
          }
        }
      } catch (e) {
        console.error("Failed to parse websocket data", e);
      }
    };

    return () => {
      ws.close();
      Object.values(camTimeoutsRef.current).forEach(clearTimeout);
    };
  }, []);

  // ===== SYNC CANVAS SIZE AND DRAW BBOX =====
  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const draw = () => {
      if (!ctx || !canvas || !container) return;

      // Sync internal attributes with physical display size
      canvas.width = container.clientWidth;
      canvas.height = container.clientHeight;

      // Clear canvas
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      if (detections.length === 0) return;

      // Kalkulasi offset letterbox/pillarbox karena iframe (webRTC player)
      // melakukan 'object-fit: contain' bawaan saat masuk mode Fullscreen
      const frameW = detections[0].frame_size?.w || 1920;
      const frameH = detections[0].frame_size?.h || 1080;
      const videoAspect = frameW / frameH;
      const canvasAspect = canvas.width / canvas.height;

      let drawW = canvas.width;
      let drawH = canvas.height;
      let offsetX = 0;
      let offsetY = 0;

      // Pencocokan rasio layar vs video
      if (canvasAspect > videoAspect) {
        // Layar lebih lebar dari video (pillarbox) -> ada pita hitam di sisi Kiri dan Kanan
        drawH = canvas.height;
        drawW = canvas.height * videoAspect;
        offsetX = (canvas.width - drawW) / 2;
      } else {
        // Layar lebih tinggi dari video (letterbox) -> ada pita hitam di sisi Atas dan Bawah
        drawW = canvas.width;
        drawH = canvas.width / videoAspect;
        offsetY = (canvas.height - drawH) / 2;
      }

      // 🔥 ITERASI SEMUA DETECTIONS SESUAI REFERENSI
      detections.forEach((d) => {
        if (!d.bbox || !d.frame_size) return;

        const { x1, y1, x2, y2 } = d.bbox;

        // SCALING PROSES berdasarkan dimensi video aktif, BUKAN dimensi fullscreen HP
        const scaleX = drawW / (d.frame_size.w || 1);
        const scaleY = drawH / (d.frame_size.h || 1);

        const sx1 = offsetX + x1 * scaleX;
        const sy1 = offsetY + y1 * scaleY;
        const sx2 = offsetX + x2 * scaleX;
        const sy2 = offsetY + y2 * scaleY;

        // WARNA SESUAI REFERENSI (LIME & RED)
        const color = d.name !== "Unknown" ? "lime" : "red";

        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        ctx.lineJoin = "round";

        // Draw box
        ctx.strokeRect(sx1, sy1, sx2 - sx1, sy2 - sy1);

        // Draw label background
        ctx.fillStyle = color;
        const label = `${d.name} (${d.person_id})`;
        ctx.font = "bold 14px Inter, Arial, sans-serif";
        const textWidth = ctx.measureText(label).width;

        ctx.fillRect(sx1, sy1 > 25 ? sy1 - 25 : sy1, textWidth + 10, 25);

        // Draw label text
        ctx.fillStyle = "#000000"; // Black text for better contrast on lime/red
        ctx.fillText(label, sx1 + 5, sy1 > 25 ? sy1 - 8 : sy1 + 18);
      });
    };

    // Add ResizeObserver to handle window/container resizing
    const resizeObserver = new ResizeObserver(() => {
      draw();
    });
    resizeObserver.observe(container);

    // Initial draw
    draw();

    return () => {
      resizeObserver.disconnect();
    };
  }, [detections]);

  const isCamOnline = isConnected && camStatus[selectedCam] !== false;

  return (
    <Card className={cn("flex flex-col overflow-hidden pb-[2px]", className)}>
      <CardHeader>
        <div className="flex items-center gap-1">
          <div className="grow">
            <CameraSelect value={selectedCam} onChange={handleCamChange} />
          </div>

          <CardAction className="flex items-center pt-2">
            <Badge
              variant={isCamOnline ? "outline" : "destructive"}
              className={cn(
                "flex items-center gap-1",
                isCamOnline && "border-green-500/30 bg-green-500/10 text-green-500",
              )}
            >
              <Circle
                className={cn("size-2 animate-pulse fill-current", isCamOnline ? "text-green-500" : "text-destructive")}
              />
              {isCamOnline ? "Online" : "Offline"}
            </Badge>
          </CardAction>
        </div>
      </CardHeader>

      <CardContent className="relative flex-grow overflow-hidden bg-black p-0">
        <div className="relative h-full w-full">
          <AspectRatio ratio={16 / 9} className="group overflow-hidden">
            <div
              ref={containerRef}
              style={{
                border: "1px solid #222",
              }}
              className="absolute inset-0 z-10 h-full w-full"
            >
              <iframe
                src={CAMERAS.find((c) => c.value === selectedCam)?.url || CAMERAS[0].url}
                className="h-full w-full border-none grayscale-[0.1] transition-all group-hover:grayscale-0"
                title="CCTV Stream"
                allow="autoplay"
              />

              <button
                type="button"
                onClick={toggleFullscreen}
                className="absolute right-2 top-2 z-[200] rounded-md bg-black/40 p-1.5 text-white/70 opacity-0 backdrop-blur-md transition-all hover:bg-black/80 hover:text-white group-hover:opacity-100"
                aria-label="Toggle Fullscreen"
              >
                {isFullscreen ? <Minimize className="size-4" /> : <Maximize className="size-4" />}
              </button>

              <canvas
                ref={canvasRef}
                className="pointer-events-none absolute inset-0 z-[100] h-full w-full"
                style={{ width: "100%", height: "100%", display: "block" }}
              />

              {!isCamOnline && (
                <div className="absolute inset-0 z-[120] flex flex-col items-center justify-center bg-black/80 backdrop-blur-sm">
                  <ShieldAlert className="mb-2 size-10 text-destructive" />
                  <p className="font-bold text-base text-white">Connection Lost</p>
                  <p className="text-[10px] text-zinc-400">Unable to reach the camera server</p>
                </div>
              )}
            </div>
          </AspectRatio>
        </div>
      </CardContent>
    </Card>
  );
}

function CameraSelect({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" role="combobox" aria-expanded={open} className="w-[180px] justify-between">
          <div className="flex items-center gap-2">
            <div
              className="size-2 rounded-full bg-primary"
              style={{
                boxShadow: "0 0 8px color-mix(in oklab, var(--primary) 50%, transparent)",
              }}
            />
            {CAMERAS.find((cam) => cam.value === value)?.label}
          </div>
          <ChevronsUpDown className="opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[180px] p-0">
        <Command>
          <CommandList>
            <CommandGroup>
              {CAMERAS.map((cam) => (
                <CommandItem
                  key={cam.value}
                  value={cam.value}
                  onSelect={(currentValue) => {
                    onChange(currentValue);
                    setOpen(false);
                  }}
                >
                  <div className="flex flex-col">
                    <span>{cam.label}</span>
                    <span className="text-muted-foreground text-xs">{cam.description}</span>
                  </div>
                  <Check className={cn("ml-auto", value === cam.value ? "opacity-100" : "opacity-0")} />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
