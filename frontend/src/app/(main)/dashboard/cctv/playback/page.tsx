"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";

export const dynamic = "force-dynamic";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { cctvApi, streamFileUrl, type PublicCamera, type RecordingItem, type Segment } from "@/lib/cctv-api";
import { Timeline } from "../_components/timeline";

function PlaybackInner() {
  const sp = useSearchParams();
  const [cams, setCams] = useState<PublicCamera[]>([]);
  const [camId, setCamId] = useState(sp.get("camera") ?? "cam1");
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [recs, setRecs] = useState<RecordingItem[]>([]);
  const [segments, setSegments] = useState<Segment[]>([]);
  const [current, setCurrent] = useState("");
  const [events, setEvents] = useState<{ person_id: number; action: string; at: string }[]>([]);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    cctvApi.cameras().then((l) => {
      setCams(l);
      if (!l.find((c) => c.id === camId) && l[0]) setCamId(l[0].id);
    }).catch(() => {});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function load() {
    setErr("");
    setLoading(true);
    try {
      const t = await cctvApi.timeline(camId, date);
      setSegments(t.segments ?? []);
      setRecs(t.recordings ?? []);
      const first = (t.segments ?? [])[0];
      const rec = (t.recordings ?? [])[0];
      if (first && rec) {
        setCurrent(streamFileUrl(rec.id, first.file));
        cctvApi.events(rec.id).then(setEvents).catch(() => setEvents([]));
      } else {
        setCurrent("");
        setEvents([]);
      }
    } catch (e) {
      setErr(e instanceof Error ? e.message : "gagal memuat");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [camId, date]);

  return (
    <div className="flex flex-col gap-4">
      <h1 className="font-bold text-xl">CCTV — Playback</h1>
      <div className="flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm">
          Kamera
          <select value={camId} onChange={(e) => setCamId(e.target.value)} className="rounded border bg-background px-2 py-1.5">
            {cams.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name || c.id}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Tanggal
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} className="w-[180px]" />
        </label>
        <Button onClick={load} size="sm" disabled={loading}>
          {loading ? "Memuat…" : "Muat"}
        </Button>
      </div>
      {loading ? <p className="text-muted-foreground text-sm">Memuat timeline…</p> : null}
      {err ? <p className="text-destructive text-sm">{err}</p> : null}
      <Card>
        <CardHeader>
          <CardTitle>
            {camId} — {date}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {current ? (
            <video key={current} src={current} controls className="max-h-[440px] w-full rounded bg-black" preload="metadata" />
          ) : (
            <p className="text-muted-foreground text-sm">Tidak ada rekaman pada tanggal ini (retensi 7 hari).</p>
          )}
          <Timeline
            segments={segments}
            onPick={(s) => {
              const rec = recs[0];
              if (rec) {
                setCurrent(streamFileUrl(rec.id, s.file));
                cctvApi.events(rec.id).then(setEvents).catch(() => setEvents([]));
              }
            }}
          />
          {events.length > 0 ? (
            <ul className="max-h-40 overflow-auto rounded border p-2 text-xs">
              {events.map((e, i) => (
                <li key={i}>
                  {new Date(e.at).toLocaleTimeString()} — person {e.person_id} {e.action}
                </li>
              ))}
            </ul>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}

export default function PlaybackPage() {
  return (
    <Suspense>
      <PlaybackInner />
    </Suspense>
  );
}
