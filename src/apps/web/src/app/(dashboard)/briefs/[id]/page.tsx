import Link from "next/link";
import { notFound } from "next/navigation";
import { createServerClient } from "@/lib/trpc/server";
import { BriefEditor } from "./_components/brief-editor";

export default async function BriefDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const client = await createServerClient();

  let brief: Awaited<ReturnType<typeof client.briefs.get>>;
  try {
    brief = await client.briefs.get({ briefId: id });
  } catch {
    notFound();
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8 sm:px-8 lg:px-12">
      <Link href="/briefs" className="text-sm text-white/65 hover:text-white">
        Back to briefs
      </Link>

      <BriefEditor
        briefId={brief.briefId}
        productUrl={brief.productUrl}
        status={brief.status}
        createdAt={brief.createdAt}
        updatedAt={brief.updatedAt}
        angles={brief.angles}
        hooks={brief.hooks}
      />
    </main>
  );
}
