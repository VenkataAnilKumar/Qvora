"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

type AssetSummary = {
  id: string;
  angle: string;
  status: string;
};

interface ExportPanelClientProps {
  assetIds: string[];
  selectedAssets: AssetSummary[];
}

const PLATFORMS = [
  { id: "tiktok", label: "TikTok (9:16)" },
  { id: "youtube_shorts", label: "YouTube Shorts (9:16)" },
  { id: "meta", label: "Meta Feed" },
] as const;

export function ExportPanelClient({ assetIds, selectedAssets }: ExportPanelClientProps) {
  const router = useRouter();
  const [platform, setPlatform] = useState<(typeof PLATFORMS)[number]["id"]>("tiktok");
  const [error, setError] = useState<string | null>(null);
  const [ready, setReady] = useState<{ downloadUrl: string } | null>(null);

  const count = assetIds.length;
  const exampleAngle = selectedAssets[0]?.angle || "angle";
  const namingPreview = useMemo(() => {
    const slug = (s: string) =>
      s
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/(^-|-$)/g, "");
    const today = new Date().toISOString().slice(0, 10).replace(/-/g, "");
    return `brand_${slug(exampleAngle)}_${platform}_${today}_v1.mp4`;
  }, [exampleAngle, platform]);

  const createExport = trpc.exports.create.useMutation({
    onSuccess: (result) => {
      setError(null);
      setReady({ downloadUrl: result.downloadUrl });
    },
    onError: (err) => setError(err.message),
  });

  const handleDownload = () => {
    if (ready) {
      window.location.href = ready.downloadUrl;
      return;
    }
    setError(null);
    createExport.mutate({ variantIds: assetIds, format: "mp4_1080p", platform });
  };

  return (
    <div className="space-y-6">
      <button
        type="button"
        onClick={() => router.push("/assets")}
        className="w-fit text-sm text-white/65 transition-colors hover:text-white"
      >
        ← Back to assets
      </button>

      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-white/40">Export</p>
        <h1 className="text-3xl font-semibold tracking-[-0.03em]">
          Export {count} asset{count !== 1 ? "s" : ""}
        </h1>
      </div>

      {count === 0 ? (
        <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
          No assets selected. Go back to the asset library and select at least one asset to export.
        </p>
      ) : (
        <>
          <Card className="border-white/8 bg-white/[0.03]">
            <CardHeader>
              <CardTitle className="text-xl">Format</CardTitle>
            </CardHeader>
            <CardContent className="flex gap-3">
              <div className="flex flex-1 items-center justify-between rounded-xl border border-[var(--color-volt)]/50 bg-[var(--color-volt)]/8 px-4 py-3">
                <span className="text-sm font-medium text-white">MP4 (H.264, 1080p)</span>
                <Badge variant="outline">default</Badge>
              </div>
              <div
                className="flex flex-1 cursor-not-allowed items-center justify-between rounded-xl border border-white/8 bg-black/20 px-4 py-3 opacity-40"
                title="4K export needs a Rust transcode pass that doesn't exist yet"
              >
                <span className="text-sm font-medium text-white">4K / GIF</span>
                <Badge variant="outline">soon</Badge>
              </div>
            </CardContent>
          </Card>

          <Card className="border-white/8 bg-white/[0.03]">
            <CardHeader>
              <CardTitle className="text-xl">Platform</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-wrap gap-3">
              {PLATFORMS.map((p) => (
                <label
                  key={p.id}
                  className="flex cursor-pointer items-center gap-2 rounded-xl border border-white/8 bg-black/20 px-4 py-2 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
                >
                  <input
                    type="radio"
                    name="platform"
                    checked={platform === p.id}
                    onChange={() => {
                      setPlatform(p.id);
                      setReady(null);
                    }}
                    className="h-4 w-4 accent-[var(--color-volt)]"
                  />
                  <span className="text-sm text-white">{p.label}</span>
                </label>
              ))}
            </CardContent>
          </Card>

          <Card className="border-white/8 bg-white/[0.03]">
            <CardHeader>
              <CardTitle className="text-xl">Naming convention</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <p className="rounded-xl border border-white/8 bg-black/30 px-4 py-3 font-mono text-xs text-white/70">
                {namingPreview}
              </p>
              <div className="flex items-center gap-2 text-sm text-white/60">
                <Badge variant="outline">✓</Badge>
                manifest.csv included
              </div>
            </CardContent>
          </Card>

          {platform === "meta" && (
            <p className="text-xs text-white/40">
              Note: Meta export doesn't reframe to 1:1 yet — you'll get the same 9:16 master asset
              as TikTok/YouTube Shorts.
            </p>
          )}

          {error && (
            <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
              {error}
            </p>
          )}

          {ready && (
            <p className="rounded-xl border border-[var(--color-convert-green)]/30 bg-[var(--color-convert-green)]/8 px-4 py-3 text-sm text-[var(--color-convert-green)]">
              ✓ Download ready
            </p>
          )}

          <div className="flex items-center gap-3 border-t border-white/8 pt-4">
            <Button disabled={createExport.isPending} className="flex-1" onClick={handleDownload}>
              {createExport.isPending
                ? "Bundling…"
                : ready
                  ? "Download ZIP"
                  : "Generate & Download ZIP"}
            </Button>
            <Button variant="outline" onClick={() => router.push("/assets")}>
              Cancel
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
