"use client";

import { useSession } from "../hooks/use-session";
import { useLogout } from "../hooks/use-logout";
import { Button } from "@/components/ui/button";

export function UserMenu() {
  const { user } = useSession();
  const { logout, pending } = useLogout();

  return (
    <div className="flex items-center gap-3">
      {user && <span className="text-muted-foreground text-sm">{user.email}</span>}
      <Button variant="outline" size="sm" onClick={() => void logout()} disabled={pending}>
        {pending ? "Signing out…" : "Sign out"}
      </Button>
    </div>
  );
}
