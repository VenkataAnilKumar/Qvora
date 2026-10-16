"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle } from "@qvora/ui";

type AssetSummary = {
  id: string;
  angle: string;
  status: string;
};

interface ExportPanelClientProps {
  assetIds: string[];
  selectedAssets: AssetSummary[];
}

export function ExportPanelClient({ assetIds, selectedAssets }: ExportPanelClientProps) {
  const router = useRouter();
  const [format, setFormat] = useState<"mp4" | "mov">("mp4");
  const [resolution, setResolution] = useState<"1080p" | "720p">("1080p");

  const count = assetIds.length;
  const exampleAngle = selectedAssets[0]?.angle || "angle";
  const exampleExt = format === "mp4" ? "mp4" : "mov";
  const namingPreview = useMemo(() => {
    const slug = (s: string) =>
      s
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/(^-|-$)/g, "");
    const today = new Date().toISOString().slice(0, 10).replace(/-/g, "");
    return `brand_${slug(exampleAngle)}_${resolution}_${today}_v1.${exampleExt}`;
  }, [exampleAngle, resolution, exampleExt]);

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
              {(["mp4", "mov"] as const).map((f) => (
                <label
                  key={f}
                  className="flex flex-1 cursor-pointer items-center justify-between rounded-xl border border-white/8 bg-black/20 px-4 py-3 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
                >
                  <span className="text-sm font-medium text-white">
                    {f === "mp4" ? "MP4 (H.264)" : "MOV"}
                  </span>
                  <input
                    type="radio"
                    name="format"
                    checked={format === f}
                    onChange={() => setFormat(f)}
                    className="h-4 w-4 accent-[var(--color-volt)]"
                  />
                </label>
              ))}
            </CardContent>
          </Card>

          <Card className="border-white/8 bg-white/[0.03]">
            <CardHeader>
              <CardTitle className="text-xl">Resolution</CardTitle>
            </CardHeader>
            <CardContent className="flex gap-3">
              {(["1080p", "720p"] as const).map((r) => (
                <label
                  key={r}
                  className="flex flex-1 cursor-pointer items-center justify-between rounded-xl border border-white/8 bg-black/20 px-4 py-3 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/8"
                >
                  <span className="text-sm font-medium text-white">{r}</span>
                  <input
                    type="radio"
                    name="resolution"
                    checked={resolution === r}
                    onChange={() => setResolution(r)}
                    className="h-4 w-4 accent-[var(--color-volt)]"
                  />
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

          <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
            Export generation isn't wired up on the backend yet — only listing existing exports is
            implemented server-side. This panel is ready to go once that endpoint exists.
          </p>

          <div className="flex items-center gap-3 border-t border-white/8 pt-4">
            <Button
              disabled
              className="flex-1"
              title="Backend export-creation endpoint not built yet"
            >
              Download ZIP
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
