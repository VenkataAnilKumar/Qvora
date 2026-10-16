"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import posthog from "posthog-js";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle } from "@qvora/ui";
import { PLAN_LIMITS } from "@qvora/types";
import type { PlanTier, VideoModel } from "@qvora/types";
import { trpc } from "@/lib/trpc/client";

type Angle = { angle: string; headline: string; script: string; cta: string; voiceTone: string };
type Brief = {
  briefId: string;
  productUrl: string;
  status: string;
  angles: Angle[];
  hooks: string[];
};

const MODELS: { id: VideoModel; label: string }[] = [
  { id: "veo3", label: "Veo 3.1" },
  { id: "kling3", label: "Kling 3.0" },
  { id: "runway4", label: "Runway Gen-4.5" },
  { id: "sora2", label: "Sora 2" },
];

const PLATFORMS = ["meta", "tiktok", "both"] as const;
const DURATIONS = [15, 30, 60] as const;
const NEXT_TIER: Record<PlanTier, PlanTier | null> = {
  starter: "growth",
  growth: "agency",
  agency: null,
};

export function GenerateSettingsClient({
  brief,
  planTier,
  usedVariants,
}: {
  brief: Brief;
  planTier: PlanTier;
  usedVariants: number;
}) {
  const router = useRouter();
  const maxPerAngle = PLAN_LIMITS[planTier].maxVariantsPerAngle;

  const [model, setModel] = useState<VideoModel>("veo3");
  const [variantsPerAngle, setVariantsPerAngle] = useState(1);
  // Platform/duration don't map to a backend field yet (batchGenerate only takes angle/hook/model) —
  // they only drive the client-side time estimate below.
  const [platform, setPlatform] = useState<(typeof PLATFORMS)[number]>("both");
  const [duration, setDuration] = useState<(typeof DURATIONS)[number]>(30);

  const [error, setError] = useState<string | null>(null);
  const [limitExceeded, setLimitExceeded] = useState(false);
  const nextTier = NEXT_TIER[planTier];

  const checkoutMutation = trpc.workspace.createCheckoutSession.useMutation({
    onSuccess: (result) => {
      posthog.capture("checkout_started", { plan_tier: nextTier });
      window.location.href = result.url;
    },
    onError: (err) => setError(err.message),
  });

  const batchGenerate = trpc.briefs.batchGenerate.useMutation({
    onSuccess: (result) => {
      setError(null);
      setLimitExceeded(false);
      posthog.capture("generation_started", {
        brief_id: brief.briefId,
        model,
        variants_per_angle: variantsPerAngle,
        jobs_created: result.jobs.length,
      });
      router.push(`/briefs/${brief.briefId}`);
    },
    onError: (err) => {
      if (err.data?.code === "PAYMENT_REQUIRED") {
        setLimitExceeded(true);
      } else {
        setError(err.message);
      }
    },
  });

  const totalVideos = brief.angles.length * variantsPerAngle * (platform === "both" ? 2 : 1);
  const estimatedMinutes = Math.max(1, Math.round(totalVideos * (duration / 30) * 1.5));

  const exceedsLimit = maxPerAngle !== null && variantsPerAngle > maxPerAngle;

  const handleGenerate = () => {
    if (batchGenerate.isPending || exceedsLimit) return;
    setError(null);
    setLimitExceeded(false);
    batchGenerate.mutate({
      briefId: brief.briefId,
      variantsPerSpec: variantsPerAngle,
      specs: brief.angles.map((angle, i) => ({
        angle: angle.angle,
        hook: brief.hooks[i] ?? brief.hooks[0] ?? angle.headline,
        model,
      })),
    });
  };

  return (
    <div className="space-y-6">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-white/40">Generate</p>
        <h1 className="text-3xl font-semibold tracking-[-0.03em]">{brief.productUrl}</h1>
        <p className="mt-1 text-sm text-white/55">
          {brief.angles.length} angle{brief.angles.length !== 1 ? "s" : ""} approved ·{" "}
          {variantsPerAngle} video{variantsPerAngle !== 1 ? "s" : ""} per angle
        </p>
      </div>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Model</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 sm:grid-cols-2">
          {MODELS.map((m) => (
            <label
              key={m.id}
              className="flex cursor-pointer items-center justify-between rounded-xl border border-white/8 bg-black/20 px-4 py-3 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
            >
              <span className="text-sm font-medium text-white">{m.label}</span>
              <input
                type="radio"
                name="model"
                checked={model === m.id}
                onChange={() => setModel(m.id)}
                className="h-4 w-4 accent-[var(--color-volt)]"
              />
            </label>
          ))}
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Variants per angle</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          {[1, 2, 3, 5, 10].map((n) => (
            <label
              key={n}
              className="flex cursor-pointer items-center gap-2 rounded-xl border border-white/8 bg-black/20 px-4 py-2 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
            >
              <input
                type="radio"
                name="variants"
                checked={variantsPerAngle === n}
                onChange={() => setVariantsPerAngle(n)}
                className="h-4 w-4 accent-[var(--color-volt)]"
              />
              <span className="text-sm text-white">{n}</span>
            </label>
          ))}
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Platform</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          {PLATFORMS.map((p) => (
            <label
              key={p}
              className="flex cursor-pointer items-center gap-2 rounded-xl border border-white/8 bg-black/20 px-4 py-2 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
            >
              <input
                type="radio"
                name="platform"
                checked={platform === p}
                onChange={() => setPlatform(p)}
                className="h-4 w-4 accent-[var(--color-volt)]"
              />
              <span className="text-sm capitalize text-white">
                {p === "meta" ? "Meta Feed (1:1)" : p === "tiktok" ? "TikTok (9:16)" : "Both"}
              </span>
            </label>
          ))}
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Duration</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          {DURATIONS.map((d) => (
            <label
              key={d}
              className="flex cursor-pointer items-center gap-2 rounded-xl border border-white/8 bg-black/20 px-4 py-2 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
            >
              <input
                type="radio"
                name="duration"
                checked={duration === d}
                onChange={() => setDuration(d)}
                className="h-4 w-4 accent-[var(--color-volt)]"
              />
              <span className="text-sm text-white">{d}s</span>
            </label>
          ))}
        </CardContent>
      </Card>

      <p className="text-sm text-white/55">
        Estimated time: ~{estimatedMinutes} minute{estimatedMinutes !== 1 ? "s" : ""} ·{" "}
        {totalVideos} videos
      </p>

      {(exceedsLimit || limitExceeded) && (
        <div className="space-y-3 rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/5 px-4 py-4">
          <div>
            <p className="text-sm font-medium text-[var(--color-signal-red)]">
              ⚠ {variantsPerAngle} variants/angle exceeds your {planTier} plan limit
              {maxPerAngle !== null ? ` of ${maxPerAngle}` : ""}.
            </p>
            <p className="mt-1 text-sm text-white/60">
              Used so far: {usedVariants} variants. Upgrade to generate more per angle.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Badge>{planTier}</Badge>
            {nextTier && (
              <Button
                type="button"
                disabled={checkoutMutation.isPending}
                onClick={() =>
                  checkoutMutation.mutate({
                    planTier: nextTier,
                    returnPath: `/briefs/${brief.briefId}/generate`,
                  })
                }
                size="sm"
              >
                {checkoutMutation.isPending ? "Redirecting…" : `Upgrade to ${nextTier} →`}
              </Button>
            )}
            <button
              type="button"
              onClick={() => setVariantsPerAngle(maxPerAngle ?? 1)}
              className="text-sm text-white/40 hover:text-white/60"
            >
              Use plan limit instead
            </button>
          </div>
        </div>
      )}

      {error && (
        <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
          {error}
        </p>
      )}

      <div className="flex items-center justify-between border-t border-white/8 pt-4">
        <Link href={`/briefs/${brief.briefId}`} className="text-sm text-white/65 hover:text-white">
          ← Back to brief
        </Link>
        <Button disabled={batchGenerate.isPending || exceedsLimit} onClick={handleGenerate}>
          {batchGenerate.isPending ? "Generating…" : `Generate ${totalVideos} videos →`}
        </Button>
      </div>
    </div>
  );
}
