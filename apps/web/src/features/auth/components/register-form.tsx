"use client";

import { useRegister } from "../hooks/use-register";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function RegisterForm() {
  const { form, onSubmit, pending, errors } = useRegister();
  const { register } = form;

  return (
    <form
      onSubmit={(e) => void onSubmit(e)}
      className="flex w-full max-w-sm flex-col gap-4"
      noValidate
    >
      <p className="text-muted-foreground text-sm">
        Create an account. You will need to verify your email before signing in.
      </p>

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
          autoComplete="new-password"
          {...register("password")}
        />
        {errors.password && (
          <p className="text-destructive text-xs">{errors.password.message}</p>
        )}
      </div>

      <Button type="submit" disabled={pending}>
        {pending ? "Creating account…" : "Create account"}
      </Button>
    </form>
  );
}
