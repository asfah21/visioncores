"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

import { AspectRatio } from "@/components/ui/aspect-ratio";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cctvApi, type PublicCamera } from "@/lib/cctv-api";
import { CAMERAS } from "../default/_components/generated-cameras";

const GO2RTC = process.env.NEXT_PUBLIC_GO2RTC_URL ?? "http://10.10.11.5:1984";

export default function CctvLivePage() {
  const [cams, setCams] = useState<PublicCamera[]>([]);
  const [err, setErr] = useState("");

  useEffect(() => {
    cctvApi
      .cameras()
      .then(setCams)
      .catch((e) => setErr(e instanceof Error ? e.message : "gagal memuat kamera"));
  }, []);

  const list = cams.length > 0 ? cams : CAMERAS.map((c) => ({ id: c.value, name: c.label, description: c.description } as PublicCamera));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-xl">CCTV — Live</h1>
        <div className="flex gap-2">
          <Link href="/dashboard/cctv/playback" className="text-sm underline">
            Playback
          </Link>
          <Link href="/dashboard/cctv/cameras" className="text-sm underline">
            Kelola Kamera
          </Link>
        </div>
      </div>
      {err ? <p className="text-destructive text-sm">API: {err} — tampilkan fallback generated-cameras.</p> : null}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {list.map((c) => (
          <Card key={c.id} className="overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>{c.name || c.id}</CardTitle>
              <Badge variant="outline">{c.id}</Badge>
            </CardHeader>
            <CardContent className="bg-black p-0">
              <AspectRatio ratio={16 / 9}>
                <iframe src={`${GO2RTC}/stream.html?src=${c.id}`} className="h-full w-full border-none" title={`Live ${c.id}`} allow="autoplay" />
              </AspectRatio>
            </CardContent>
            <div className="flex justify-between p-3 text-sm">
              <span className="text-muted-foreground">{c.description ?? ""}</span>
              <Link href={`/dashboard/cctv/${c.id}`} className="underline">
                Detail → Live / PTZ / Playback
              </Link>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
