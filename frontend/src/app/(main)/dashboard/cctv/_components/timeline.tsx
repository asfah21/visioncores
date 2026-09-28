"use client";

import { cn } from "@/lib/utils";
import type { Segment } from "@/lib/cctv-api";

export function Timeline({ segments, onPick }: { segments: Segment[]; onPick: (s: Segment) => void }) {
  const hours = Array.from({ length: 24 }, (_, h) => h);
  const byHour = new Map<number, Segment[]>();
  for (const s of segments) {
    const h = new Date(s.start).getHours();
    if (!byHour.has(h)) byHour.set(h, []);
    byHour.get(h)?.push(s);
  }
  return (
    <div>
      <div className="flex justify-between text-[10px] text-muted-foreground">
        {["00", "03", "06", "09", "12", "15", "18", "21", "24"].map((t) => (
          <span key={t}>{t}</span>
        ))}
      </div>
      <div className="grid gap-[2px]" style={{ gridTemplateColumns: "repeat(24, minmax(0, 1fr))" }}>
        {hours.map((h) => {
          const segs = byHour.get(h) ?? [];
          return (
            <button
              key={h}
              type="button"
              title={`${String(h).padStart(2, "0")}:00 — ${segs.length} segmen`}
              onClick={() => segs[0] && onPick(segs[0])}
              className={cn(
                "h-6 rounded-sm border",
                segs.length > 0 ? "border-green-500/40 bg-green-500/40 hover:bg-green-500/60" : "border-muted bg-muted/40",
              )}
            />
          );
        })}
      </div>
      <p className="mt-1 text-[11px] text-muted-foreground">Blok hijau = ada rekaman. Klik untuk memutar segmen pertama jam itu.</p>
    </div>
  );
}
