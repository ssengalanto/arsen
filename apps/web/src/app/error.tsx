"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/error";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  const message =
    error instanceof ApiError
      ? (error.detail ?? error.title ?? "Something went wrong.")
      : "Something went wrong.";

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
      <h1 className="text-2xl font-semibold">Something went wrong</h1>
      <p className="text-muted-foreground max-w-md text-sm">{message}</p>
      <Button onClick={reset}>Try again</Button>
    </div>
  );
}
