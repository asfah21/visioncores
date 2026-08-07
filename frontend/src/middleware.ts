import { type NextRequest, NextResponse } from "next/server";

import { SESSION_COOKIE, verifySessionToken } from "@/lib/auth";

// Routes that are publicly accessible (no login required)
const PUBLIC_PATHS = ["/auth/v1/login", "/auth/v2/login", "/auth/v1/register", "/auth/v2/register"];

// Routes that logged-in users should NOT access (redirect to dashboard)
const AUTH_ONLY_PATHS = ["/auth"];

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const sessionToken = request.cookies.get(SESSION_COOKIE)?.value;
  const isAuthenticated = sessionToken ? !!(await verifySessionToken(sessionToken)) : false;

  // ── Root path → redirect based on auth status ─────────────────────────────
  if (pathname === "/") {
    if (isAuthenticated) {
      return NextResponse.redirect(new URL("/dashboard/default", request.url));
    }
    return NextResponse.redirect(new URL("/auth/v1/login", request.url));
  }

  // ── Authenticated users trying to access auth pages → go to dashboard ─────
  if (isAuthenticated && AUTH_ONLY_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.redirect(new URL("/dashboard/default", request.url));
  }

  // ── Unauthenticated users trying to access protected routes ───────────────
  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));
  if (!isAuthenticated && !isPublic) {
    // Allow static assets and Next.js internals to pass through
    if (pathname.startsWith("/_next") || pathname.startsWith("/api") || pathname.includes(".")) {
      return NextResponse.next();
    }
    return NextResponse.redirect(new URL("/auth/v1/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  // Run middleware on all routes except static files
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
