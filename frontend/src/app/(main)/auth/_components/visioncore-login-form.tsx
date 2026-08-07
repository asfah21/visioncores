"use client";

import { useActionState, useEffect, useState } from "react";

import { Eye, EyeOff, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { loginAction } from "@/server/auth-actions";

export function VisionCoreLoginForm() {
  const [state, formAction, isPending] = useActionState(loginAction, null);
  const [showPassword, setShowPassword] = useState(false);

  // Show error toast when server returns error
  useEffect(() => {
    if (state?.error) {
      toast.error(state.error, {
        id: "login-error",
        duration: 4000,
      });
    }
  }, [state]);

  return (
    <div className="flex flex-col gap-6">
      <form action={formAction} className="flex flex-col gap-2" noValidate>
        {/* Email/Username field */}
        <div className="mb-1 flex flex-col gap-2">
          <label className="px-0.5 font-medium text-foreground/90 text-sm" htmlFor="vc-username">
            Username or Email
          </label>
          <div className="relative flex items-center">
            <input
              id="vc-username"
              name="username"
              type="text"
              className="w-full rounded-lg border border-input bg-background px-4 py-2.5 text-[15px] text-foreground outline-none transition-all placeholder:text-muted-foreground/50 focus:border-primary focus:ring-2 focus:ring-primary/20 [&:-webkit-autofill]:text-fill-foreground [&:-webkit-autofill]:shadow-[0_0_0_1000px_inset_var(--background)]"
              placeholder="admin"
              autoComplete="username"
              // biome-ignore lint/a11y/noAutofocus: intended for login page
              autoFocus
              required
              disabled={isPending}
            />
          </div>
        </div>

        {/* Password field */}
        <div className="mb-2 flex flex-col gap-2">
          <label className="px-0.5 font-medium text-foreground/90 text-sm" htmlFor="vc-password">
            Password
          </label>
          <div className="relative flex items-center">
            <input
              id="vc-password"
              name="password"
              type={showPassword ? "text" : "password"}
              className="w-full rounded-lg border border-input bg-background px-4 py-2.5 text-[15px] text-foreground outline-none transition-all placeholder:text-muted-foreground/50 focus:border-primary focus:ring-2 focus:ring-primary/20 [&:-webkit-autofill]:text-fill-foreground [&:-webkit-autofill]:shadow-[0_0_0_1000px_inset_var(--background)]"
              placeholder="••••••••"
              autoComplete="current-password"
              required
              disabled={isPending}
            />
            <button
              type="button"
              className="absolute right-3 text-muted-foreground/60 transition-colors hover:text-foreground"
              onClick={() => setShowPassword((p) => !p)}
              aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
              tabIndex={-1}
            >
              {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>

        {/* Remember me */}
        <div className="mb-6 flex items-center gap-2.5 px-0.5">
          <input
            type="checkbox"
            id="remember"
            name="remember"
            className="h-4 w-4 rounded border-input bg-background text-primary focus:ring-primary/30"
          />
          <label htmlFor="remember" className="cursor-pointer select-none font-medium text-muted-foreground/80 text-sm">
            Remember me
          </label>
        </div>

        {/* Error message inline */}
        {state?.error && (
          <div
            className="fade-in zoom-in-95 mb-4 animate-in rounded-lg border border-destructive/20 bg-destructive/10 p-3 font-medium text-destructive text-sm"
            role="alert"
          >
            {state.error}
          </div>
        )}

        {/* Submit */}
        <button
          id="vc-login-submit"
          type="submit"
          className="flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2.5 font-semibold text-primary-foreground shadow-sm transition-all hover:opacity-90 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
          disabled={isPending}
        >
          {isPending ? (
            <>
              <Loader2 size={18} className="animate-spin" />
              <span>Logging in...</span>
            </>
          ) : (
            <span>Login</span>
          )}
        </button>
      </form>

      <div className="relative my-1 text-center text-sm">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-border/60 border-t" />
        </div>
        <span className="relative z-10 bg-background px-3 text-muted-foreground/60">
          PT. Gunung Samudera Internasional
        </span>
      </div>

      {/* Google Button
      <button 
        className="flex items-center justify-center gap-3 w-full py-2.5 px-4 bg-background hover:bg-muted/30 border border-input text-foreground font-semibold rounded-lg transition-all active:scale-[0.98]"
      >
        <svg className="h-5 w-5" viewBox="0 0 24 24">
          <path
            d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
            fill="#4285F4"
          />
          <path
            d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
            fill="#34A853"
          />
          <path
            d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
            fill="#FBBC05"
          />
          <path
            d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 12-4.53z"
            fill="#EA4335"
          />
        </svg>
        <span>Continue with Google</span>
      </button> */}
    </div>
  );
}
