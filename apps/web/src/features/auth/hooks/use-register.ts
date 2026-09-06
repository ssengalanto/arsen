"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api/error";
import { registerSchema, type RegisterInput } from "../schemas/register";
import { register as registerUser } from "../fetchers";

export function useRegister() {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  const form = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: "", password: "" },
  });
  const { errors } = form.formState;

  const onSubmit = form.handleSubmit(async (values) => {
    setPending(true);
    try {
      await registerUser(values);
      router.push("/login?registered=1");
    } catch (err) {
      if (err instanceof ApiError) {
        for (const [field, message] of Object.entries(err.fieldErrors())) {
          form.setError(field as keyof RegisterInput, { message });
        }
        if (err.errors.length === 0) {
          form.setError("root", {
            message: "Registration failed. Please try again.",
          });
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
