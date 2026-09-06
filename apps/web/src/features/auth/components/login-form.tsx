"use client";

import { useLogin } from "../hooks/use-login";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function LoginForm({ registered = false }: { registered?: boolean }) {
  const { form, onSubmit, pending, errors } = useLogin();
  const { register } = form;

  return (
    <form
      onSubmit={(e) => void onSubmit(e)}
      className="flex w-full max-w-sm flex-col gap-4"
      noValidate
    >
      {registered && (
        <p className="text-muted-foreground rounded-md border p-3 text-sm" role="status">
          Registration successful. Check your email to verify your account, then sign in.
        </p>
      )}

      {errors.root && (
        <p className="text-destructive text-sm" role="alert">
          {errors.root.message}
        </p>
      )}

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="email">Email</Label>
        <Input id="email" type="email" autoComplete="email" {...register("email")} />
        {errors.email && <p className="text-destructive text-xs">{errors.email.message}</p>}
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          type="password"
          autoComplete="current-password"
          {...register("password")}
        />
        {errors.password && (
          <p className="text-destructive text-xs">{errors.password.message}</p>
        )}
      </div>

      <Button type="submit" disabled={pending}>
        {pending ? "Signing in…" : "Sign in"}
      </Button>
    </form>
  );
}
