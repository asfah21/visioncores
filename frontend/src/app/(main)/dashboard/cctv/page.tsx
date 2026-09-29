"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Video } from "lucide-react";

import { AspectRatio } from "@/components/ui/aspect-ratio";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cctvApi, type PublicCamera } from "@/lib/cctv-api";
import { AddCameraDialog } from "./_components/add-camera-dialog";

export default function CctvPage() {
  const [cams, setCams] = useState<PublicCamera[]>([]);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);

  async function load() {
    setLoading(true);
    setErr("");
    try {
      setCams((await cctvApi.cameras()) ?? []);
    } catch (e) {
      setCams([]);
      setErr(e instanceof Error ? e.message : "Failed to load cameras.");
    } finally {
      setLoading(false);
    }
  }

  // biome-ignore lint/correctness/useExhaustiveDependencies: load on mount only; refreshes via onAdded
  useEffect(() => {
    load();
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-xl">CCTV</h1>
        <Button size="sm" onClick={() => setDialogOpen(true)}>
          <Plus className="mr-1 size-4" /> Add Camera
        </Button>
      </div>
      {err ? <p className="text-destructive text-sm">{err}</p> : null}
      {loading ? (
        <p className="text-muted-foreground text-sm">Loading cameras…</p>
      ) : cams.length === 0 && !err ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-2 py-10 text-center">
            <Video className="size-8 text-muted-foreground" />
            <p className="font-medium">No cameras yet</p>
            <p className="text-muted-foreground text-sm">Click “Add Camera” to register your first camera.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {cams.map((c) => (
            <Link key={c.id} href={`/dashboard/cctv/${c.id}`} className="block">
              <Card className="overflow-hidden transition-colors hover:border-primary">
                <div className="bg-black p-0">
                  <AspectRatio ratio={16 / 9}>
                    <div className="flex h-full w-full flex-col items-center justify-center gap-1 bg-muted/30 text-muted-foreground">
                      <Video className="size-8" />
                      <span className="text-xs">Click to view live</span>
                    </div>
                  </AspectRatio>
                </div>
                <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2">
                  <CardTitle className="truncate text-base">{c.name || c.id}</CardTitle>
                  {c.enabled ? <Badge variant="outline">Active</Badge> : <Badge variant="secondary">Disabled</Badge>}
                </CardHeader>
                <CardContent className="pt-0 text-muted-foreground text-sm">
                  <p className="truncate">{c.description || c.id}</p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
      <AddCameraDialog open={dialogOpen} onOpenChange={setDialogOpen} onAdded={load} />
    </div>
  );
}
