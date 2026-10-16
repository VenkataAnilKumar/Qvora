import { createServerClient } from "@/lib/trpc/server";
import { AssetLibraryClient } from "./_components/asset-library-client";

export const metadata = { title: "Assets — Qvora" };

export default async function AssetsPage() {
  const client = await createServerClient();

  let initialData: Awaited<ReturnType<typeof client.assets.list>> | null = null;
  let initialError: string | null = null;

  try {
    initialData = await client.assets.list();
  } catch (err) {
    initialError = err instanceof Error ? err.message : "Failed to load assets";
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-7xl flex-1 flex-col gap-8 px-6 py-8 sm:px-8 lg:px-12">
      <AssetLibraryClient initialData={initialData} initialError={initialError} />
    </main>
  );
}
