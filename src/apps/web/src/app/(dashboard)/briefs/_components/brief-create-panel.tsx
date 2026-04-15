"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle, Input } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

type Recommendation = {
  angle: string;
  suggested_hook: string;
  rationale: string;
  confidence_score: number;
  impression_volume: number;
  window_days: number;
  last_generated_at: string;
};

interface BriefCreatePanelProps {
  initialProductUrl?: string;
  initialTemplate?: string;
  recommendations: Recommendation[];
}

function formatAngleLabel(angle: string): string {
  return angle
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function BriefCreatePanel({
  initialProductUrl,
  initialTemplate,
  recommendations,
}: BriefCreatePanelProps) {
  const router = useRouter();
  const [productUrl, setProductUrl] = useState(initialProductUrl ?? "");
  const [template, setTemplate] = useState(initialTemplate ?? "");
  const [ignoredAngles, setIgnoredAngles] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  const feedbackMutation = trpc.signal.submitRecommendationFeedback.useMutation();

  const visibleRecommendations = useMemo(
    () => recommendations.filter((item) => !ignoredAngles.includes(item.angle)),
    [recommendations, ignoredAngles],
  );

  const activeRecommendation = useMemo(
    () => visibleRecommendations.find((item) => item.angle === template) ?? null,
    [visibleRecommendations, template],
  );

  const createMutation = trpc.briefs.create.useMutation({
    onSuccess: (result) => {
      setError(null);
      router.push(`/briefs/${result.briefId}`);
    },
    onError: (mutationError) => {
      setError(mutationError.message);
    },
  });

  return (
    <Card className="border-white/8 bg-white/[0.03]">
      <CardHeader>
        <CardTitle className="text-xl">Create new brief</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <label
            htmlFor="brief-product-url"
            className="text-xs uppercase tracking-[0.16em] text-white/45"
          >
            Product URL
          </label>
          <Input
            id="brief-product-url"
            type="url"
            value={productUrl}
            onChange={(e) => setProductUrl(e.target.value)}
            placeholder="https://brand.com/products/hero-offer"
            disabled={createMutation.isPending}
          />
        </div>

        <div className="space-y-2">
          <label
            htmlFor="brief-template"
            className="text-xs uppercase tracking-[0.16em] text-white/45"
          >
            Strategy hint (optional)
          </label>
          <Input
            id="brief-template"
            value={template}
            onChange={(e) => setTemplate(e.target.value)}
            placeholder="problem_solution"
            disabled={createMutation.isPending}
          />
          {activeRecommendation ? (
            <p className="text-xs text-white/60">
              Suggested hook: {activeRecommendation.suggested_hook}
            </p>
          ) : null}
        </div>

        {visibleRecommendations.length > 0 ? (
          <div className="space-y-2 rounded-2xl border border-white/8 bg-black/20 p-4">
            <p className="text-xs uppercase tracking-[0.16em] text-white/45">
              Signal recommendations
            </p>
            <div className="grid gap-2 md:grid-cols-3">
              {visibleRecommendations.map((item) => (
                <div
                  key={item.angle}
                  className="rounded-xl border border-white/10 bg-white/[0.03] px-3 py-2 text-left"
                >
                  <button
                    type="button"
                    className="w-full text-left transition hover:text-[var(--color-volt)]"
                    onClick={() => {
                      setTemplate(item.angle);
                      feedbackMutation.mutate({
                        angle: item.angle,
                        action: "accept",
                        source: "brief_create_panel",
                      });
                    }}
                    disabled={createMutation.isPending || feedbackMutation.isPending}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <p className="text-sm font-medium text-white">
                        {formatAngleLabel(item.angle)}
                      </p>
                      <Badge variant="outline">{item.confidence_score.toFixed(1)}%</Badge>
                    </div>
                    <p className="mt-1 text-xs text-white/55">
                      {item.impression_volume.toLocaleString()} impressions
                    </p>
                  </button>

                  <div className="mt-2 flex items-center justify-end">
                    <button
                      type="button"
                      className="text-xs uppercase tracking-[0.14em] text-white/45 transition hover:text-white"
                      onClick={() => {
                        feedbackMutation.mutate({
                          angle: item.angle,
                          action: "ignore",
                          source: "brief_create_panel",
                        });
                        setIgnoredAngles((prev) => [...prev, item.angle]);
                        if (template === item.angle) {
                          setTemplate("");
                        }
                      }}
                      disabled={createMutation.isPending || feedbackMutation.isPending}
                    >
                      Ignore
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {error ? (
          <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/10 px-3 py-2 text-sm text-[var(--color-signal-red)]">
            {error}
          </p>
        ) : null}

        <div className="flex items-center justify-between gap-3">
          <p className="text-xs text-white/50">
            Recommendations are generated from workspace signal data.
          </p>
          <Button
            type="button"
            disabled={createMutation.isPending || productUrl.trim().length === 0}
            onClick={() => {
              if (createMutation.isPending) return;
              setError(null);
              createMutation.mutate({
                productUrl: productUrl.trim(),
                template: template.trim() || undefined,
                model: "gpt-4o",
              });
            }}
          >
            {createMutation.isPending ? "Generating…" : "Generate brief"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
