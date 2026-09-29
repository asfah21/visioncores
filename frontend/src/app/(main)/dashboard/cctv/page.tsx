"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Cctv, Pencil, Plus, Trash2, Video } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { CAMERA_BRANDS, cctvApi, type PublicCamera } from "@/lib/cctv-api";
import { CameraDialog } from "./_components/camera-dialog";
import { DeleteCameraDialog } from "./_components/delete-camera-dialog";

function brandLabel(brand?: string) {
  return CAMERA_BRANDS.find((b) => b.value === (brand ?? ""))?.label ?? "Generic ONVIF";
}

export default function CctvPage() {
  const [cams, setCams] = useState<PublicCamera[]>([]);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(true);
  const [addOpen, setAddOpen] = useState(false);
  const [editing, setEditing] = useState<PublicCamera | null>(null);
  const [deleting, setDeleting] = useState<PublicCamera | null>(null);

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

  // biome-ignore lint/correctness/useExhaustiveDependencies: load on mount only; refreshes after add/edit/delete
  useEffect(() => {
    load();
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-xl">CCTV</h1>
        <Button size="sm" onClick={() => setAddOpen(true)}>
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
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
          {cams.map((c) => (
            <Card key={c.id} className="transition-colors hover:border-primary">
              <Link href={`/dashboard/cctv/${c.id}`} className="block">
                <div className="flex items-start gap-3 p-4 pb-2">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                    <Cctv className="size-6" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[11px] tracking-wider text-muted-foreground uppercase">
                      {brandLabel(c.brand)}
                    </p>
                    <p className="truncate font-semibold leading-tight">{c.name || c.id}</p>
                    {c.description ? <p className="truncate text-muted-foreground text-xs">{c.description}</p> : null}
                  </div>
                  {c.enabled ? (
                    <Badge variant="outline" className="shrink-0">
                      Active
                    </Badge>
                  ) : (
                    <Badge variant="secondary" className="shrink-0">
                      Disabled
                    </Badge>
                  )}
                </div>
              </Link>
              <div className="flex justify-end gap-1 px-3 pb-2">
                <Button
                  size="sm"
                  variant="ghost"
                  aria-label={`Edit ${c.id}`}
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setEditing(c);
                  }}
                >
                  <Pencil className="size-4" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  aria-label={`Delete ${c.id}`}
                  className="text-destructive hover:text-destructive"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setDeleting(c);
                  }}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
      <CameraDialog open={addOpen} onOpenChange={setAddOpen} onSaved={load} />
      <CameraDialog open={!!editing} onOpenChange={(v) => !v && setEditing(null)} onSaved={load} camera={editing} />
      <DeleteCameraDialog camera={deleting} onOpenChange={(v) => !v && setDeleting(null)} onDeleted={load} />
    </div>
  );
}
