import { SWRConfig } from "swr";
import { resourcesKey } from "@/lib/swr/keys";
import { listResources } from "@/lib/resource-store";
import { ResourceList } from "@/features/resource";

// RSC entry for the worked slice. The initial list is read on the server and
// handed to SWR as `fallback`, so the client paints real data immediately and
// then revalidates through the BFF.
export default function ResourcesPage() {
  const initial = listResources();

  return (
    <SWRConfig value={{ fallback: { [resourcesKey]: initial } }}>
      <section className="flex flex-col gap-6">
        <div className="flex flex-col gap-1">
          <h1 className="font-heading text-xl font-semibold">Resources</h1>
          <p className="text-muted-foreground text-sm">
            A worked CRUD slice: SWR for server data, Zustand for UI state, and the BFF proxy.
          </p>
        </div>
        <ResourceList />
      </section>
    </SWRConfig>
  );
}
