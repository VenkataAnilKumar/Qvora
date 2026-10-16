import { notFound } from "next/navigation";
import { createServerClient } from "@/lib/trpc/server";
import { GenerateSettingsClient } from "./_components/generate-settings-client";

export default async function GenerateSettingsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const client = await createServerClient();

  let brief: Awaited<ReturnType<typeof client.briefs.get>>;
  let workspace: Awaited<ReturnType<typeof client.workspace.get>>;
  let usage: Awaited<ReturnType<typeof client.workspace.getUsage>>;
  try {
    [brief, workspace, usage] = await Promise.all([
      client.briefs.get({ briefId: id }),
      client.workspace.get(),
      client.workspace.getUsage(),
    ]);
  } catch {
    notFound();
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-3xl flex-1 flex-col gap-8 px-6 py-8 sm:px-8 lg:px-12">
      <GenerateSettingsClient
        brief={brief}
        planTier={workspace.planTier}
        usedVariants={usage.usedVariants}
      />
    </main>
  );
}
