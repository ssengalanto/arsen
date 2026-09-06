"use client";

import type { ReactNode } from "react";
import { SWRConfig } from "swr";
import { Toaster } from "@/components/ui/sonner";
import { swrConfig } from "@/lib/swr/config";

export function Providers({ children }: { children: ReactNode }) {
  return (
    <SWRConfig value={swrConfig}>
      {children}
      <Toaster richColors position="top-right" />
    </SWRConfig>
  );
}
