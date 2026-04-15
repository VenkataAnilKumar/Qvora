import { Badge, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@qvora/ui";
import { PLAN_LIMITS } from "@qvora/types";
import { createServerClient } from "@/lib/trpc/server";
import { DashboardClient } from "./_components/dashboard-client";

const pipeline = [
  "Playwright scrape via Modal",
  "GPT-4o structured brief generation",
  "FAL queue submission and render orchestration",
  "Rust postprocess and Mux delivery",
] as const;

export default async function DashboardPage() {
  const client = await createServerClient();
  const [workspace, jobs] = await Promise.all([
    client.workspace.get(),
    client.generation.listJobs({ limit: 20 }),
  ]);
  const planLimit = PLAN_LIMITS[workspace.planTier].maxVariantsPerAngle;
  const planLimitLabel = planLimit === null ? "Unlimited" : `${planLimit} / angle`;
  const initialJobs = jobs.jobs.map((j) => ({
    jobId: j.jobId,
    productUrl: j.productUrl,
    status: j.status,
  }));

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-7xl flex-1 flex-col gap-8 px-6 py-8 sm:px-8 lg:px-12">
      {/* ── Hero / stats section — server-rendered ── */}
      <section className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Card className="border-white/10 bg-[linear-gradient(135deg,rgba(123,47,255,0.16),rgba(17,17,24,0.98)_48%,rgba(46,156,255,0.10))]">
          <CardHeader className="gap-4">
            <div className="flex flex-wrap gap-2">
              <Badge>{workspace.planTier}</Badge>
              <Badge variant="success">{workspace.subscriptionStatus}</Badge>
              <Badge variant="outline">Workspace {workspace.orgId}</Badge>
            </div>
            <div className="space-y-3">
              <CardDescription>Generation console</CardDescription>
              <CardTitle className="text-4xl tracking-[-0.03em]">
                Turn one landing page into a test-ready video batch.
              </CardTitle>
            </div>
            <p className="max-w-2xl text-sm leading-7 text-white/65">
              Qvora keeps the V1 loop tight: ingest one product URL, pick the rendering model, and
              hand the rest to the async generation pipeline. Limits stay server-enforced by plan
              tier.
            </p>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-3">
            <StatCard label="Plan limit" value={planLimitLabel} />
            <StatCard label="Queued jobs" value={String(jobs.jobs.length)} />
            <StatCard label="Pipeline mode" value="Async" />
          </CardContent>
        </Card>

        <Card className="border-white/8 bg-white/[0.03]">
          <CardHeader>
            <CardDescription>Current runbook</CardDescription>
            <CardTitle className="text-2xl">Generation path</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm text-white/65">
            {pipeline.map((item, index) => (
              <div
                key={item}
                className="flex items-start gap-3 rounded-2xl border border-white/8 bg-black/20 px-4 py-3"
              >
                <span className="mt-0.5 inline-flex h-6 w-6 items-center justify-center rounded-full bg-[var(--color-volt)]/15 text-xs font-semibold text-[var(--color-volt)]">
                  {index + 1}
                </span>
                <span>{item}</span>
              </div>
            ))}
          </CardContent>
        </Card>
      </section>

      {/* ── Form + jobs list — client-rendered ── */}
      <DashboardClient planLimitLabel={planLimitLabel} initialJobs={initialJobs} />
    </main>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-white/8 bg-black/20 p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-white/40">{label}</p>
      <p className="mt-3 text-2xl font-semibold text-white">{value}</p>
    </div>
  );
}
