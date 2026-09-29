"use client";

import { useState } from "react";
import { TriangleAlert } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { cctvApi, type PublicCamera } from "@/lib/cctv-api";

export function DeleteCameraDialog({
  camera,
  onOpenChange,
  onDeleted,
}: {
  camera: PublicCamera | null;
  onOpenChange: (v: boolean) => void;
  onDeleted: () => void;
}) {
  const [msg, setMsg] = useState("");
  const [deleting, setDeleting] = useState(false);

  async function confirm() {
    if (!camera) return;
    setMsg("");
    setDeleting(true);
    try {
      await cctvApi.deleteCamera(camera.id);
      try {
        await cctvApi.go2rtcSync();
      } catch {
        // non-fatal: record is deleted, stream drops on next sync/restart
      }
      onOpenChange(false);
      onDeleted();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "Failed to delete camera.");
    } finally {
      setDeleting(false);
    }
  }

  return (
    <Dialog open={!!camera} onOpenChange={(v) => !deleting && onOpenChange(v)}>
      <DialogContent className="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <TriangleAlert className="size-5 text-destructive" /> Delete Camera
          </DialogTitle>
          <DialogDescription>
            Delete <span className="font-semibold text-foreground">{camera?.name || camera?.id}</span> ({camera?.id})? Recordings already
            stored on disk are kept, but the camera, its live stream, and future recordings will be removed. This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        {msg ? <p className="text-sm text-amber-500">{msg}</p> : null}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={deleting}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={confirm} disabled={deleting}>
            {deleting ? "Deleting…" : "Delete"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
