"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter, useSearchParams } from "next/navigation";
import { ApiError } from "@/lib/api/error";
import { loginSchema, type LoginInput } from "../schemas/login";
import { login } from "../fetchers";
import { useAuthStore } from "../store";

export function useLogin() {
  const router = useRouter();
  const params = useSearchParams();
  const setUser = useAuthStore((s) => s.setUser);
  const [pending, setPending] = useState(false);

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });
  const { errors } = form.formState;

  const onSubmit = form.handleSubmit(async (values) => {
    setPending(true);
    try {
      const { user } = await login(values);
      setUser(user);
      router.push(params.get("next") ?? "/dashboard");
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        form.setError("root", {
          message: "Please verify your email before signing in.",
        });
        return;
      }
      if (err instanceof ApiError) {
        for (const [field, message] of Object.entries(err.fieldErrors())) {
          form.setError(field as keyof LoginInput, { message });
        }
        if (err.errors.length === 0) {
          form.setError("root", { message: "Login failed. Check your credentials." });
        }
        return;
      }
      form.setError("root", { message: "Something went wrong. Please try again." });
    } finally {
      setPending(false);
    }
  });

  return { form, onSubmit, pending, errors };
}
