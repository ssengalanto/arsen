import Link from "next/link";
import { notFound } from "next/navigation";
import { SWRConfig, unstable_serialize } from "swr";
import { resourceKey } from "@/lib/swr/keys";
import { getResource } from "@/lib/resource-store";
import { ResourceDetail } from "@/features/resource";

// RSC detail entry. The resource is read on the server and seeded into SWR's
// `fallback` under the serialized tuple key so `useResource(id)` hydrates
// without a client roundtrip.
export default async function ResourceDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const resource = getResource(id);
  if (!resource) notFound();

  return (
    <SWRConfig value={{ fallback: { [unstable_serialize(resourceKey(id))]: resource } }}>
      <section className="flex flex-col gap-6">
        <Link href="/resources" className="text-muted-foreground text-sm hover:text-foreground">
          ← Back to resources
        </Link>
        <ResourceDetail id={id} />
      </section>
    </SWRConfig>
  );
}
