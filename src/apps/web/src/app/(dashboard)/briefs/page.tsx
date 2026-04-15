import Link from "next/link";
import { Badge, Card, CardContent, CardHeader, CardTitle } from "@qvora/ui";
import { createServerClient } from "@/lib/trpc/server";

export default async function BriefsPage() {
  const client = await createServerClient();
  const briefs = await client.briefs.list();

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-6xl flex-1 flex-col gap-6 px-6 py-8 sm:px-8 lg:px-12">
      <header className="space-y-2">
        <p className="text-xs uppercase tracking-[0.2em] text-white/40">Creative briefing</p>
        <h1 className="text-3xl font-semibold tracking-[-0.03em]">Brief history</h1>
      </header>

      {briefs.briefs.length === 0 ? (
        <Card className="border-white/8 bg-white/[0.03]">
          <CardContent className="py-10 text-center text-white/60">
            No briefs yet. Generate your first brief from a product URL.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {briefs.briefs.map((brief) => (
            <Link key={brief.briefId} href={`/briefs/${brief.briefId}`}>
              <Card className="border-white/8 bg-white/[0.03] transition hover:border-[var(--color-volt)]/40">
                <CardHeader className="flex flex-row items-start justify-between gap-4">
                  <div className="space-y-2">
                    <CardTitle className="text-base leading-6">{brief.productUrl}</CardTitle>
                    <p className="text-xs text-white/45">
                      {new Date(brief.createdAt).toLocaleString()}
                    </p>
                  </div>
                  <Badge variant={brief.status === "generated" ? "success" : "outline"}>
                    {brief.status}
                  </Badge>
                </CardHeader>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </main>
  );
}
