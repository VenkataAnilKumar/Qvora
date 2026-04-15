import { Badge, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@qvora/ui";
import { PLAN_LIMITS } from "@qvora/types";
import { createServerClient } from "@/lib/trpc/server";
import { DashboardClient } from "./_components/dashboard-client";

type TrendPoint = {
  date: string;
  total: number;
  accepted: number;
  acceptance_rate: number;
};

const pipeline = [
  "Playwright scrape via Modal",
  "GPT-4o structured brief generation",
  "FAL queue submission and render orchestration",
  "Rust postprocess and Mux delivery",
] as const;

export default async function DashboardPage() {
  const client = await createServerClient();
  const [workspace, usage, jobs, signalDashboard, signalConnections, fatigue, feedbackByAngle] =
    await Promise.all([
      client.workspace.get(),
      client.workspace.getUsage(),
      client.generation.listJobs({ limit: 20 }),
      client.signal.getDashboard({ days: 30 }),
      client.signal.listConnections(),
      client.signal.detectFatigue({
        days: 30,
        dropPct: 30,
        sustainedDays: 3,
        minPeakCtr: 0.01,
        minImpressions: 1000,
        persist: true,
      }),
      client.signal.listRecommendationFeedbackByAngle({ days: 30, limit: 5 }),
    ]);
  const planLimit = PLAN_LIMITS[workspace.planTier].maxVariantsPerAngle;
  const planLimitLabel = planLimit === null ? "Unlimited" : `${planLimit} / angle`;
  const utilizationPercent =
    planLimit === null || planLimit <= 0
      ? null
      : Math.min(100, Math.round((usage.usedVariants / planLimit) * 100));
  const utilizationWidthClass =
    utilizationPercent === null
      ? "w-full"
      : utilizationPercent <= 0
        ? "w-0"
        : utilizationPercent <= 10
          ? "w-[10%]"
          : utilizationPercent <= 20
            ? "w-[20%]"
            : utilizationPercent <= 30
              ? "w-[30%]"
              : utilizationPercent <= 40
                ? "w-[40%]"
                : utilizationPercent <= 50
                  ? "w-[50%]"
                  : utilizationPercent <= 60
                    ? "w-[60%]"
                    : utilizationPercent <= 70
                      ? "w-[70%]"
                      : utilizationPercent <= 80
                        ? "w-[80%]"
                        : utilizationPercent <= 90
                          ? "w-[90%]"
                          : "w-full";
  const usageLabel =
    planLimit === null
      ? `${usage.usedVariants} used (unlimited)`
      : `${usage.usedVariants} / ${planLimit} used`;
  const connectedCount = signalConnections.connections.filter(
    (c) => c.status === "connected",
  ).length;
  const signalCtrLabel = `${signalDashboard.totals.ctr.toFixed(2)}%`;
  const signalRoasLabel = signalDashboard.totals.roas.toFixed(2);
  const topAngle = signalDashboard.by_angle[0]?.angle ?? "No data yet";
  const fatigueCount = fatigue.alerts.length;
  const topFatigueDrop = fatigue.alerts[0]?.drop_pct;
  const topFatigueLabel =
    topFatigueDrop === undefined ? "No active fatigue" : `${topFatigueDrop.toFixed(1)}% drop`;
  const feedbackCount = signalDashboard.recommendation_feedback.total;
  const feedbackAcceptanceRate = `${signalDashboard.recommendation_feedback.acceptance_rate.toFixed(1)}%`;
  const acceptanceDelta = signalDashboard.recommendation_feedback.acceptance_delta_pct_points;
  const acceptanceDeltaLabel = `${acceptanceDelta >= 0 ? "+" : ""}${acceptanceDelta.toFixed(1)} pp`;
  const acceptanceDeltaTone =
    acceptanceDelta >= 0 ? "text-[var(--color-convert-green)]" : "text-[var(--color-signal-red)]";
  const acceptanceTrendPath = buildAcceptanceTrendPath(
    signalDashboard.recommendation_feedback.trend,
  );
  const topAcceptedAngles = feedbackByAngle.feedback_by_angle
    .filter((item) => item.accepted > 0)
    .sort((a, b) => b.accepted - a.accepted || b.acceptance_rate - a.acceptance_rate)
    .slice(0, 3);
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
            <StatCard label="Usage" value={usageLabel} />
            <StatCard label="Queued jobs" value={String(jobs.jobs.length)} />
            <StatCard
              label="Utilization"
              value={utilizationPercent === null ? "N/A" : `${utilizationPercent}%`}
            />
          </CardContent>
          <CardContent className="pt-0">
            <div className="rounded-2xl border border-white/8 bg-black/20 p-4">
              <div className="mb-2 flex items-center justify-between text-xs uppercase tracking-[0.16em] text-white/45">
                <span>Usage this cycle</span>
                <span>{usageLabel}</span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-white/10">
                <div
                  className={`h-full bg-[var(--color-data-blue)] transition-all ${utilizationWidthClass}`}
                />
              </div>
              <p className="mt-2 text-xs text-white/45">
                Last reset{" "}
                {usage.lastResetAt ? new Date(usage.lastResetAt).toLocaleString() : "not available"}
              </p>
            </div>
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

        <Card className="border-white/8 bg-white/[0.03]">
          <CardHeader>
            <CardDescription>Qvora Signal (Phase 2 kickoff)</CardDescription>
            <CardTitle className="text-2xl">Performance snapshot</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm text-white/65">
            <div className="grid gap-3 sm:grid-cols-2">
              <StatCard label="CTR (30d)" value={signalCtrLabel} />
              <StatCard label="ROAS (30d)" value={signalRoasLabel} />
              <StatCard label="Connected accounts" value={String(connectedCount)} />
              <StatCard label="Top angle" value={topAngle} />
              <StatCard label="Fatigue alerts" value={String(fatigueCount)} />
              <StatCard label="Worst fatigue" value={topFatigueLabel} />
              <StatCard label="Recommendation feedback" value={String(feedbackCount)} />
              <StatCard label="Recommendation acceptance" value={feedbackAcceptanceRate} />
            </div>
            <div className="rounded-2xl border border-white/8 bg-black/20 p-4">
              <div className="flex items-center justify-between gap-3">
                <p className="text-xs uppercase tracking-[0.16em] text-white/45">
                  Acceptance trend (last 14 days)
                </p>
                <p className={`text-xs font-semibold ${acceptanceDeltaTone}`}>
                  7d delta {acceptanceDeltaLabel}
                </p>
              </div>
              <svg
                viewBox="0 0 100 32"
                className="mt-3 h-16 w-full"
                preserveAspectRatio="none"
                aria-hidden="true"
              >
                <defs>
                  <linearGradient id="acceptanceTrendStroke" x1="0" y1="0" x2="1" y2="0">
                    <stop offset="0%" stopColor="var(--color-data-blue)" stopOpacity="0.7" />
                    <stop offset="100%" stopColor="var(--color-convert-green)" stopOpacity="0.95" />
                  </linearGradient>
                </defs>
                <path
                  d={acceptanceTrendPath}
                  fill="none"
                  stroke="url(#acceptanceTrendStroke)"
                  strokeWidth="2.2"
                  strokeLinecap="round"
                />
              </svg>
              <p className="mt-2 text-xs text-white/45">
                Current 7d acceptance{" "}
                {signalDashboard.recommendation_feedback.current_7d_rate.toFixed(1)}% vs previous 7d{" "}
                {signalDashboard.recommendation_feedback.previous_7d_rate.toFixed(1)}%.
              </p>
            </div>
            {topAcceptedAngles.length > 0 ? (
              <div className="rounded-2xl border border-white/8 bg-black/20 p-4">
                <p className="mb-3 text-xs uppercase tracking-[0.16em] text-white/45">
                  Top accepted angles (30d)
                </p>
                <div className="space-y-2">
                  {topAcceptedAngles.map((item, index) => (
                    <div key={item.angle} className="flex items-center justify-between text-xs">
                      <span className="text-white/70">
                        {index + 1}. {item.angle}
                      </span>
                      <span className="font-medium text-[var(--color-convert-green)]">
                        {item.accepted} accepts
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ) : null}
            <p className="text-xs text-white/45">
              Metrics are sourced from Signal ingestion endpoints and grouped by angle/platform for
              next-gen brief learning.
            </p>
            {fatigueCount > 0 ? (
              <div className="rounded-2xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/10 p-4 text-xs text-[var(--color-qvora-white)]">
                <p className="font-semibold text-[var(--color-signal-red)]">
                  Active fatigue detected
                </p>
                <p className="mt-1 text-white/80">
                  {fatigue.alerts[0]?.angle ?? "unknown"} variant{" "}
                  {fatigue.alerts[0]?.variant_id ?? "n/a"} has dropped{" "}
                  {fatigue.alerts[0]?.drop_pct?.toFixed(1) ?? "0.0"}% from its recent peak.
                  Suggested action: {fatigue.alerts[0]?.suggested_action ?? "refresh_hook"}.
                </p>
              </div>
            ) : null}
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

function buildAcceptanceTrendPath(points: TrendPoint[]): string {
  if (points.length === 0) {
    return "M 2 16 L 98 16";
  }

  const width = 96;
  const height = 24;
  const xOffset = 2;
  const yOffset = 4;

  const maxRate = Math.max(...points.map((point) => point.acceptance_rate), 1);
  const minRate = Math.min(...points.map((point) => point.acceptance_rate), 0);
  const range = Math.max(maxRate - minRate, 1);

  return points
    .map((point, index) => {
      const x = xOffset + (index / Math.max(points.length - 1, 1)) * width;
      const normalized = (point.acceptance_rate - minRate) / range;
      const y = yOffset + (1 - normalized) * height;
      return `${index === 0 ? "M" : "L"} ${x.toFixed(2)} ${y.toFixed(2)}`;
    })
    .join(" ");
}
