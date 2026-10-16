import { createServerClient } from "@/lib/trpc/server";
import { BrandKitClient } from "./_components/brand-kit-client";

export const metadata = { title: "Brand Kit — Qvora" };

export default async function BrandPage() {
  const client = await createServerClient();

  let brandKit: Awaited<ReturnType<typeof client.workspace.getBrandKit>> | null = null;
  try {
    brandKit = await client.workspace.getBrandKit();
  } catch {
    // Non-fatal — form renders with defaults below.
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-2xl flex-1 flex-col gap-8 px-6 py-8 sm:px-8 lg:px-12">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-white/40">Workspace</p>
        <h1 className="text-3xl font-semibold tracking-[-0.03em]">Brand Kit</h1>
      </div>
      <BrandKitClient initialBrandKit={brandKit} />
    </main>
  );
}
