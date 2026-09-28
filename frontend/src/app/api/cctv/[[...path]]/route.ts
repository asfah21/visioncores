import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { SESSION_COOKIE } from "@/lib/auth";

const BACKEND = process.env.BACKEND_URL ?? "http://backend:3001";

async function forward(req: NextRequest, path: string[]) {
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value;
  if (!token) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  const qs = req.nextUrl.search ?? "";
  const url = `${BACKEND}/api/${path.join("/")}${qs}`;
  const headers = new Headers();
  const ct = req.headers.get("content-type");
  if (ct) headers.set("content-type", ct);
  headers.set("Authorization", `Bearer ${token}`);
  const range = req.headers.get("range");
  if (range) headers.set("range", range);

  const init: RequestInit = { method: req.method, headers, cache: "no-store" };
  if (req.method !== "GET" && req.method !== "HEAD") {
    const buf = await req.arrayBuffer();
    init.body = buf;
  }
  let upstream: Response;
  try {
    upstream = await fetch(url, init);
  } catch {
    return NextResponse.json({ error: "backend tidak terjangkau (cek container backend_api)" }, { status: 502 });
  }
  const outHeaders = new Headers();
  for (const k of ["content-type", "content-length", "accept-ranges", "content-range"]) {
    const v = upstream.headers.get(k);
    if (v) outHeaders.set(k, v);
  }
  if (upstream.status === 204) return new NextResponse(null, { status: 204 });
  const buf = await upstream.arrayBuffer();
  return new NextResponse(buf, { status: upstream.status, headers: outHeaders });
}

export async function GET(req: NextRequest, ctx: { params: Promise<{ path?: string[] }> }) {
  const { path = [] } = await ctx.params;
  return forward(req, path);
}
export async function POST(req: NextRequest, ctx: { params: Promise<{ path?: string[] }> }) {
  const { path = [] } = await ctx.params;
  return forward(req, path);
}
export async function PUT(req: NextRequest, ctx: { params: Promise<{ path?: string[] }> }) {
  const { path = [] } = await ctx.params;
  return forward(req, path);
}
export async function DELETE(req: NextRequest, ctx: { params: Promise<{ path?: string[] }> }) {
  const { path = [] } = await ctx.params;
  return forward(req, path);
}
