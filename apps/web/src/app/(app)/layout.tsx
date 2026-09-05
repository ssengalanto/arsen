import type { ReactNode } from "react";
import Link from "next/link";
import { UserMenu } from "@/features/auth";

// Protected group shell. The proxy already gates entry by cookie presence;
// this renders the authenticated chrome around every protected page.
export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-full flex-col">
      <header className="flex items-center justify-between border-b px-6 py-3">
        <nav className="flex items-center gap-4">
          <Link href="/dashboard" className="font-semibold">
            Arsen
          </Link>
          <Link href="/resources" className="text-muted-foreground text-sm hover:text-foreground">
            Resources
          </Link>
        </nav>
        <UserMenu />
      </header>
      <main className="flex flex-1 flex-col p-6">{children}</main>
    </div>
  );
}
