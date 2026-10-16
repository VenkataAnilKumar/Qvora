import { createServerClient } from "@/lib/trpc/server";
import { ExportPanelClient } from "./_components/export-panel-client";

export const metadata = { title: "Export — Qvora" };

export default async function ExportNewPage({
  searchParams,
}: {
  searchParams: Promise<{ assetIds?: string }>;
}) {
  const { assetIds: assetIdsParam } = await searchParams;
  const assetIds = (assetIdsParam ?? "").split(",").filter(Boolean);

  const client = await createServerClient();
  let selectedAssets: Awaited<ReturnType<typeof client.assets.list>>["assets"] = [];
  try {
    const { assets } = await client.assets.list();
    selectedAssets = assets.filter((a) => assetIds.includes(a.id));
  } catch {
    // Non-fatal — the client component shows an honest empty/error state below.
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-xl flex-1 flex-col gap-8 px-6 py-8 sm:px-8 lg:px-12">
      <ExportPanelClient assetIds={assetIds} selectedAssets={selectedAssets} />
    </main>
  );
}
