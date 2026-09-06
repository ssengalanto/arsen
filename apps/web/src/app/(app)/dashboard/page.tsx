export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      <p className="text-muted-foreground text-sm">
        You are signed in. This is the protected landing page.
      </p>
    </div>
  );
}
