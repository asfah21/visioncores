"use client";

import { use, useEffect, useState } from "react";

import { AspectRatio } from "@/components/ui/aspect-ratio";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cctvApi, streamFileUrl, type PublicCamera, type Segment } from "@/lib/cctv-api";
import { PtzPad } from "../_components/ptz-pad";
import { Timeline } from "../_components/timeline";

const GO2RTC = process.env.NEXT_PUBLIC_GO2RTC_URL ?? "http://10.10.11.5:1984";

export default function CameraDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [cam, setCam] = useState<PublicCamera | null>(null);
  const [info, setInfo] = useState("");
  const [segments, setSegments] = useState<Segment[]>([]);
  const [current, setCurrent] = useState("");
  const [recording, setRecording] = useState(false);
  const [err, setErr] = useState("");

  async function load() {
    try {
      const c = await cctvApi.camera(id);
      setCam(c);
      try {
        const i = await cctvApi.info(id);
        setInfo(`${i.device.manufacturer} ${i.device.model} (${i.device.firmware})`);
      } catch {
        setInfo("ONVIF info tidak tersedia");
      }
      const r = await cctvApi.recordings(id, `?limit=50`);
      setSegments(r.segments);
      setRecording(r.recording);
      if (r.segments[0] && !current) {
        const recs = r.recordings[0];
        if (recs) setCurrent(streamFileUrl(recs.id, r.segments[0].file));
      }
    } catch (e) {
      setErr(e instanceof Error ? e.message : "gagal memuat");
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <h1 className="font-bold text-xl">Kamera {cam?.name ?? id}</h1>
        {recording ? <Badge className="bg-red-500">● REC</Badge> : <Badge variant="outline">idle</Badge>}
      </div>
      {err ? <p className="text-destructive text-sm">{err}</p> : null}
      {info ? <p className="text-muted-foreground text-sm">{info}</p> : null}

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card className="overflow-hidden xl:col-span-2">
          <CardHeader>
            <CardTitle>Live</CardTitle>
          </CardHeader>
          <CardContent className="bg-black p-0">
            <AspectRatio ratio={16 / 9}>
              <iframe src={`${GO2RTC}/stream.html?src=${id}`} className="h-full w-full border-none" title={`Live ${id}`} allow="autoplay" />
            </AspectRatio>
          </CardContent>
          <div className="flex gap-2 p-3">
            <Button
              size="sm"
              variant={recording ? "destructive" : "default"}
              onClick={async () => {
                if (recording) await cctvApi.stopRec(id);
                else await cctvApi.startRec(id);
                load();
              }}
            >
              {recording ? "Stop Recording" : "Start Recording"}
            </Button>
          </div>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>PTZ</CardTitle>
          </CardHeader>
          <CardContent>
            <PtzPad cameraId={id} />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Playback terbaru</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {current ? (
            <video src={current} controls className="max-h-[420px] w-full rounded bg-black" preload="metadata" />
          ) : (
            <p className="text-muted-foreground text-sm">Belum ada segmen. Mulai recording untuk mengisi timeline.</p>
          )}
          <Timeline segments={segments} onPick={(s) => setCurrent(`/api/cctv/recordings/0/stream?file=${encodeURIComponent(s.file)}`)} />
        </CardContent>
      </Card>
    </div>
  );
}
