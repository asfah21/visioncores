"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import {
  createSessionToken,
  REMEMBER_ME_MAX_AGE,
  SESSION_COOKIE,
  SESSION_MAX_AGE,
  validateCredentials,
} from "@/lib/auth";

// ─── Login Action ─────────────────────────────────────────────────────────────
export async function loginAction(
  _prevState: { error?: string } | null,
  formData: FormData,
): Promise<{ error?: string }> {
  const username = (formData.get("username") as string | null)?.trim() ?? "";
  const password = (formData.get("password") as string | null) ?? "";
  const remember = formData.get("remember") === "on";

  if (!username || !password) {
    return { error: "Username dan password wajib diisi." };
  }

  const isValid = await validateCredentials(username, password);

  if (!isValid) {
    return { error: "Username atau password salah." };
  }

  // Determine session duration (8h vs 30d)
  const maxAge = remember ? REMEMBER_ME_MAX_AGE : SESSION_MAX_AGE;

  // Create signed JWT and store in secure HTTP-only cookie
  const token = await createSessionToken(username, maxAge);
  const cookieStore = await cookies();

  cookieStore.set(SESSION_COOKIE, token, {
    httpOnly: true,
    // Gunakan secure cookie hanya jika di-deploy dengan HTTPS (misalnya flag HTTPS=true)
    // karena environment internal menggunakan HTTP pada IP 10.10.11.5,
    // kita set false secara default agar cookie tidak di-reject oleh browser.
    secure: process.env.NODE_ENV === "production" && process.env.HTTPS === "true",
    sameSite: "lax",
    maxAge: maxAge,
    path: "/",
  });

  redirect("/dashboard/default");
}

// ─── Logout Action ────────────────────────────────────────────────────────────
export async function logoutAction(): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.delete(SESSION_COOKIE);
  redirect("/auth/v1/login");
}
