import Link from "next/link";
import { notFound } from "next/navigation";
import { Badge, Card, CardContent, CardHeader, CardTitle } from "@qvora/ui";
import { createServerClient } from "@/lib/trpc/server";

export default async function BriefDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const client = await createServerClient();
  const payload = await client.briefs.list();
  const brief = payload.briefs.find((item) => item.briefId === id);

  if (!brief) {
    notFound();
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8 sm:px-8 lg:px-12">
      <Link href="/briefs" className="text-sm text-white/65 hover:text-white">
        Back to briefs
      </Link>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader className="gap-3">
          <div className="flex items-center justify-between gap-4">
            <CardTitle className="text-2xl tracking-[-0.02em]">Brief {brief.briefId}</CardTitle>
            <Badge variant={brief.status === "generated" ? "success" : "outline"}>
              {brief.status}
            </Badge>
          </div>
          <p className="break-all text-sm text-white/60">{brief.productUrl}</p>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-white/65">
          <p>Created: {new Date(brief.createdAt).toLocaleString()}</p>
          <p>Scrape job: {brief.scrapeJobId}</p>
          <p>Template: {brief.template ?? "default"}</p>
        </CardContent>
      </Card>
    </main>
  );
}
