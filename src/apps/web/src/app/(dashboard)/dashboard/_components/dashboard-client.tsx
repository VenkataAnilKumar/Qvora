"use client";

import { useState } from "react";
import MuxPlayer from "@mux/mux-player-react";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
} from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";
import { JobProgress } from "./job-progress";
import type { VideoModel } from "@qvora/types";

const models: { id: VideoModel; label: string; note: string }[] = [
  { id: "veo3", label: "Veo 3.1", note: "Default for highest-quality launch variants" },
  { id: "kling3", label: "Kling 3.0", note: "Alternative motion profile for concept exploration" },
  { id: "runway4", label: "Runway Gen-4.5", note: "Fast iteration pass for creative volume" },
  { id: "sora2", label: "Sora 2", note: "Reserved for premium experimental runs" },
];

interface JobItem {
  jobId: string;
  productUrl: string;
  status: string;
}

interface DashboardClientProps {
  planLimitLabel: string;
  initialJobs: JobItem[];
}

export function DashboardClient({ planLimitLabel, initialJobs }: DashboardClientProps) {
  const utils = trpc.useUtils();
  const [url, setUrl] = useState("https://example.com/products/hero-offer");
  const [model, setModel] = useState<VideoModel>("veo3");
  const [variantId, setVariantId] = useState("");
  const [playbackLoading, setPlaybackLoading] = useState(false);
  const [playbackError, setPlaybackError] = useState<string | null>(null);
  const [playback, setPlayback] = useState<{
    variantId: string;
    playbackId: string;
    token: string;
    tokenExpires: string;
  } | null>(null);
  const [activeJob, setActiveJob] = useState<JobItem | null>(null);
  const [jobs, setJobs] = useState<JobItem[]>(initialJobs);
  const [formError, setFormError] = useState<string | null>(null);

  const submitMutation = trpc.generation.submit.useMutation({
    onSuccess: (data) => {
      const newJob: JobItem = {
        jobId: data.jobId,
        productUrl: data.productUrl,
        status: data.status,
      };
      setActiveJob(newJob);
      setJobs((prev) => [newJob, ...prev]);
      setFormError(null);
    },
    onError: (err) => {
      setFormError(err.message);
    },
  });

  const handleLoadPlayback = async () => {
    const normalizedVariantID = variantId.trim();
    if (!normalizedVariantID || playbackLoading) {
      return;
    }

    setPlaybackLoading(true);
    setPlaybackError(null);

    try {
      const data = await utils.variants.getPlaybackUrl.fetch({ variantId: normalizedVariantID });
      setPlayback({
        variantId: data.variantId,
        playbackId: data.playbackId,
        token: data.token,
        tokenExpires: data.tokenExpires,
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "failed_to_load_playback";
      setPlayback(null);
      setPlaybackError(message);
    } finally {
      setPlaybackLoading(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (submitMutation.isPending) return;
    setFormError(null);
    submitMutation.mutate({ productUrl: url, model });
  };

  return (
    <section className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
      {/* ── Generation form card ───────────────────────────────────────────── */}
      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardDescription>Create a generation run</CardDescription>
          <CardTitle className="text-2xl">Queue a new creative job</CardTitle>
        </CardHeader>
        <CardContent>
          {activeJob ? (
            <div className="space-y-4">
              <JobProgress jobId={activeJob.jobId} productUrl={activeJob.productUrl} />
              <Button
                type="button"
                variant="outline"
                className="w-full"
                onClick={() => {
                  setActiveJob(null);
                  setUrl("https://example.com/products/hero-offer");
                  setModel("veo3");
                }}
              >
                Queue another job
              </Button>
            </div>
          ) : (
            <form className="space-y-6" onSubmit={handleSubmit}>
              {/* URL input */}
              <div className="space-y-2">
                <label htmlFor="product-url" className="text-sm font-medium text-white/78">
                  Product URL
                </label>
                <Input
                  id="product-url"
                  name="product-url"
                  type="url"
                  placeholder="https://brand.com/products/offer"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  disabled={submitMutation.isPending}
                  required
                />
                <p className="text-xs text-white/45">
                  Paste the landing page or PDP you want the pipeline to scrape.
                </p>
              </div>

              {/* Model selection */}
              <div className="space-y-3">
                <p className="text-sm font-medium text-white/78">Model selection</p>
                <div className="grid gap-3 md:grid-cols-2">
                  {models.map((m) => (
                    <label
                      key={m.id}
                      className="flex cursor-pointer flex-col gap-2 rounded-2xl border border-white/8 bg-black/20 p-4 transition-colors hover:border-[var(--color-volt)]/40 hover:bg-[var(--color-volt)]/6 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div>
                          <p className="font-medium text-white">{m.label}</p>
                          <p className="text-xs uppercase tracking-[0.2em] text-white/38">{m.id}</p>
                        </div>
                        <input
                          type="radio"
                          name="model"
                          value={m.id}
                          checked={model === m.id}
                          onChange={() => setModel(m.id)}
                          disabled={submitMutation.isPending}
                          className="h-4 w-4 accent-[var(--color-volt)]"
                        />
                      </div>
                      <p className="text-sm leading-6 text-white/58">{m.note}</p>
                    </label>
                  ))}
                </div>
              </div>

              {/* Error message */}
              {formError && (
                <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
                  {formError}
                </p>
              )}

              {/* Submit row */}
              <div className="flex flex-col gap-3 border-t border-white/8 pt-4 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-white/52">Limit: {planLimitLabel}</p>
                <Button type="submit" disabled={submitMutation.isPending || !url}>
                  {submitMutation.isPending ? "Queuing…" : "Queue generation"}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>

      {/* ── Workspace jobs list card ───────────────────────────────────────── */}
      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardDescription>Recent jobs</CardDescription>
          <CardTitle className="text-2xl">Workspace queue</CardTitle>
        </CardHeader>
        <CardContent>
          {jobs.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-white/10 bg-black/20 px-5 py-8 text-sm leading-7 text-white/55">
              No generation jobs yet. Submit a URL above to start the pipeline.
            </div>
          ) : (
            <div className="space-y-3">
              {jobs.map((job) => (
                <div
                  key={job.jobId}
                  className="rounded-2xl border border-white/8 bg-black/20 px-4 py-3 text-sm text-white/64"
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="truncate font-medium text-white">{job.productUrl}</span>
                    <Badge
                      variant={
                        job.status === "complete"
                          ? "success"
                          : job.status === "failed"
                            ? "destructive"
                            : "outline"
                      }
                    >
                      {job.status}
                    </Badge>
                  </div>
                  <p className="mt-2 font-mono text-xs text-white/40">{job.jobId}</p>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03] xl:col-span-2">
        <CardHeader>
          <CardDescription>Variant playback</CardDescription>
          <CardTitle className="text-2xl">Mux stream validation</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-3 md:flex-row">
            <Input
              value={variantId}
              onChange={(e) => setVariantId(e.target.value)}
              placeholder="Enter variant UUID"
              className="md:flex-1"
            />
            <Button
              type="button"
              variant="outline"
              disabled={playbackLoading || !variantId}
              onClick={handleLoadPlayback}
            >
              {playbackLoading ? "Loading…" : "Load playback"}
            </Button>
          </div>

          {playbackError ? (
            <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
              {playbackError}
            </p>
          ) : null}

          {playback ? (
            <div className="space-y-3">
              <MuxPlayer
                playbackId={playback.playbackId}
                tokens={{ playback: playback.token }}
                metadata={{
                  video_title: `Variant ${playback.variantId}`,
                  viewer_user_id: playback.variantId,
                }}
                streamType="on-demand"
                className="aspect-[9/16] w-full max-w-sm overflow-hidden rounded-2xl border border-white/10"
              />
              <p className="text-xs text-white/45">
                Token expires at {new Date(playback.tokenExpires).toLocaleString()}.
              </p>
            </div>
          ) : (
            <p className="text-sm text-white/55">
              Enter a completed variant ID to validate signed playback delivery via Mux.
            </p>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
