"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import MuxPlayer from "@mux/mux-player-react";
import { Badge, Button, Skeleton, cn } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

type Asset = {
  id: string;
  job_id: string;
  angle: string;
  status: string;
  mux_asset_id?: string;
  mux_playback_id?: string;
  r2_key?: string;
  duration_secs?: string | number;
  created_at?: string;
  updated_at?: string;
};

type AssetListData = {
  orgId: string;
  assets: Asset[];
};

interface AssetLibraryClientProps {
  initialData: AssetListData | null;
  initialError: string | null;
}

function formatDuration(secs: string | number | undefined): string {
  if (secs == null) return "";
  const n = typeof secs === "string" ? Number.parseFloat(secs) : secs;
  if (Number.isNaN(n)) return "";
  const m = Math.floor(n / 60);
  const s = Math.round(n % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

function statusVariant(status: string): "success" | "destructive" | "outline" {
  if (status === "complete" || status === "ready") return "success";
  if (status === "failed" || status === "errored") return "destructive";
  return "outline";
}

export function AssetLibraryClient({ initialData, initialError }: AssetLibraryClientProps) {
  const router = useRouter();

  const [angleFilter, setAngleFilter] = useState("all");
  const [dateFilter, setDateFilter] = useState("all");
  // Brand/Format filters are visual-only — the assets.list response has no brand/format metadata yet.
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [previewAsset, setPreviewAsset] = useState<Asset | null>(null);

  const { data, error, isLoading } = trpc.assets.list.useQuery(undefined, {
    initialData: initialData ?? undefined,
  });

  const assets = data?.assets ?? [];

  const angles = useMemo(
    () => Array.from(new Set(assets.map((a) => a.angle).filter(Boolean))).sort(),
    [assets],
  );

  const filtered = useMemo(() => {
    let result = assets;
    if (angleFilter !== "all") {
      result = result.filter((a) => a.angle === angleFilter);
    }
    if (dateFilter !== "all") {
      const days = dateFilter === "7d" ? 7 : 30;
      const cutoff = Date.now() - days * 24 * 60 * 60 * 1000;
      result = result.filter((a) => !a.created_at || new Date(a.created_at).getTime() >= cutoff);
    }
    return result;
  }, [assets, angleFilter, dateFilter]);

  const allSelected = filtered.length > 0 && filtered.every((a) => selectedIds.has(a.id));

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleSelectAll = () => {
    setSelectedIds(allSelected ? new Set() : new Set(filtered.map((a) => a.id)));
  };

  const handleExport = () => {
    router.push(`/exports/new?assetIds=${Array.from(selectedIds).join(",")}`);
  };

  const errorMsg = error ? error.message : initialError;

  if (isLoading && !initialData) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-9 w-36" />
          <Skeleton className="h-9 w-40" />
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {["sk-1", "sk-2", "sk-3", "sk-4", "sk-5", "sk-6"].map((key) => (
            <Skeleton key={key} className="aspect-video rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-white/40">Library</p>
          <h1 className="text-3xl font-semibold tracking-[-0.03em]">Assets</h1>
        </div>
        <Button asChild>
          <Link href="/briefs">+ Generate more</Link>
        </Button>
      </div>

      {errorMsg && (
        <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
          {errorMsg}
        </p>
      )}

      {!errorMsg && (
        <>
          <div className="flex flex-wrap items-center gap-2">
            {/* Brand/Format: no backing data yet — visual-only, single "All" option */}
            <select
              disabled
              className="cursor-not-allowed rounded-lg border border-white/10 bg-black/20 px-3 py-1.5 text-sm text-white/30"
            >
              <option>Brand</option>
            </select>
            <select
              disabled
              className="cursor-not-allowed rounded-lg border border-white/10 bg-black/20 px-3 py-1.5 text-sm text-white/30"
            >
              <option>Format</option>
            </select>
            <select
              value={angleFilter}
              onChange={(e) => setAngleFilter(e.target.value)}
              className="rounded-lg border border-white/10 bg-black/20 px-3 py-1.5 text-sm text-white outline-none focus:border-[var(--color-volt)]/50"
            >
              <option value="all">Angle</option>
              {angles.map((angle) => (
                <option key={angle} value={angle}>
                  {angle}
                </option>
              ))}
            </select>
            <select
              value={dateFilter}
              onChange={(e) => setDateFilter(e.target.value)}
              className="rounded-lg border border-white/10 bg-black/20 px-3 py-1.5 text-sm text-white outline-none focus:border-[var(--color-volt)]/50"
            >
              <option value="all">Date</option>
              <option value="7d">Last 7 days</option>
              <option value="30d">Last 30 days</option>
            </select>
            {(angleFilter !== "all" || dateFilter !== "all") && (
              <button
                type="button"
                onClick={() => {
                  setAngleFilter("all");
                  setDateFilter("all");
                }}
                className="rounded-lg px-3 py-1.5 text-sm text-white/40 transition-colors hover:text-white/70"
              >
                Clear
              </button>
            )}
          </div>

          <div className="flex items-center gap-4">
            <label className="flex cursor-pointer select-none items-center gap-2">
              <input
                type="checkbox"
                checked={allSelected}
                onChange={toggleSelectAll}
                className="h-4 w-4 rounded border-white/20 accent-[var(--color-volt)]"
              />
              <span className="text-sm text-white/60">Select all</span>
            </label>
            <span className="text-sm text-white/40">
              {filtered.length} asset{filtered.length !== 1 ? "s" : ""}
            </span>
          </div>

          {filtered.length === 0 ? (
            <EmptyState />
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {filtered.map((asset) => (
                <AssetCard
                  key={asset.id}
                  asset={asset}
                  isSelected={selectedIds.has(asset.id)}
                  onToggleSelect={() => toggleSelect(asset.id)}
                  onPreview={() => setPreviewAsset(asset)}
                />
              ))}
            </div>
          )}
        </>
      )}

      {selectedIds.size > 0 && (
        <div className="fixed bottom-6 left-1/2 z-40 flex -translate-x-1/2 items-center gap-3 rounded-2xl border border-white/10 bg-black/80 px-6 py-3 shadow-2xl backdrop-blur-md">
          <span className="text-sm text-white/60">{selectedIds.size} selected</span>
          <Button onClick={handleExport}>Export selected ({selectedIds.size})</Button>
          <button
            type="button"
            onClick={() => setSelectedIds(new Set())}
            className="text-sm text-white/40 transition-colors hover:text-white/60"
          >
            Clear
          </button>
        </div>
      )}

      {previewAsset && (
        <AssetPreviewModal asset={previewAsset} onClose={() => setPreviewAsset(null)} />
      )}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <div className="mb-4 flex size-16 items-center justify-center rounded-full bg-[var(--color-volt)]/10">
        <PlayIcon className="h-7 w-7 text-[var(--color-volt)]" />
      </div>
      <h3 className="text-lg font-semibold">No ads yet</h3>
      <p className="mt-1 text-sm text-white/50">Approve a brief to generate your first set.</p>
      <Button asChild className="mt-6">
        <Link href="/briefs">Go to briefs</Link>
      </Button>
    </div>
  );
}

function AssetCard({
  asset,
  isSelected,
  onToggleSelect,
  onPreview,
}: {
  asset: Asset;
  isSelected: boolean;
  onToggleSelect: () => void;
  onPreview: () => void;
}) {
  const duration = formatDuration(asset.duration_secs);
  // No frontend download-URL mechanism was found for assets — keep disabled rather than build a broken link.
  const hasDownload = false;

  return (
    <div
      className={cn(
        "group relative overflow-hidden rounded-2xl border bg-black/20 transition-colors",
        isSelected
          ? "border-[var(--color-volt)]/50 ring-1 ring-[var(--color-volt)]/30"
          : "border-white/8 hover:border-white/20",
      )}
    >
      {/* biome-ignore lint/a11y/useSemanticElements: can't be a real <button> — it contains a nested
          interactive checkbox toggle below, and interactive elements can't nest in valid HTML */}
      <div
        role="button"
        tabIndex={0}
        className="relative flex aspect-video cursor-pointer items-center justify-center bg-white/[0.03]"
        onClick={onPreview}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onPreview();
          }
        }}
      >
        <PlayIcon className="h-10 w-10 text-white/20" />

        {duration && (
          <div className="absolute bottom-2 right-2 flex items-center gap-1 rounded-md bg-black/60 px-2 py-0.5 text-xs text-white backdrop-blur-sm">
            <PlayIcon className="h-3 w-3" filled />
            {duration}
          </div>
        )}

        <div className="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-white/20 backdrop-blur-sm">
            <PlayIcon className="h-5 w-5 text-white" filled />
          </div>
        </div>

        {/* biome-ignore lint/a11y/useSemanticElements: a real <input type="checkbox"> here would need its
            own label/positioning model distinct from this custom-styled overlay indicator; aria-checked +
            keyboard handling below already makes it accessible */}
        <div
          role="checkbox"
          aria-checked={isSelected}
          tabIndex={0}
          className={cn(
            "absolute bottom-2 left-2 transition-opacity",
            isSelected ? "opacity-100" : "opacity-0 group-hover:opacity-100",
          )}
          onClick={(e) => {
            e.stopPropagation();
            onToggleSelect();
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              e.stopPropagation();
              onToggleSelect();
            }
          }}
        >
          <div
            className={cn(
              "flex h-5 w-5 items-center justify-center rounded border-2 backdrop-blur-sm transition-colors",
              isSelected
                ? "border-[var(--color-volt)] bg-[var(--color-volt)]"
                : "border-white/60 bg-black/60",
            )}
          >
            {isSelected && (
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-3 w-3 text-white"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={3}
                aria-hidden="true"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            )}
          </div>
        </div>
      </div>

      <div className="space-y-3 p-3">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="truncate text-sm font-medium text-white">{asset.angle || "Untitled"}</p>
            {duration && <p className="text-xs text-white/40">{duration}</p>}
          </div>
          <Badge variant={statusVariant(asset.status)} className="shrink-0 capitalize">
            {asset.status}
          </Badge>
        </div>

        <div className="flex items-center gap-1.5">
          <button
            type="button"
            disabled={!hasDownload}
            title="Download (not available yet)"
            className="flex items-center gap-1 rounded-lg border border-white/10 px-2.5 py-1 text-xs text-white/60 disabled:cursor-not-allowed disabled:opacity-30"
          >
            ↓ download
          </button>
          <button
            type="button"
            disabled
            title="Edit (coming soon)"
            className="flex items-center gap-1 rounded-lg border border-white/10 px-2.5 py-1 text-xs text-white/60 disabled:cursor-not-allowed disabled:opacity-30"
          >
            ✎ edit
          </button>
          <button
            type="button"
            title="More options (coming soon)"
            className="ml-auto flex h-7 w-7 items-center justify-center rounded-lg border border-white/10 text-white/60 transition-colors hover:border-white/20 hover:text-white"
          >
            ⋯
          </button>
        </div>
      </div>
    </div>
  );
}

function AssetPreviewModal({ asset, onClose }: { asset: Asset; onClose: () => void }) {
  // Mux playback is signed — mint a token via the variants router before rendering the player,
  // matching the same pattern dashboard-client.tsx uses for variant playback validation.
  const {
    data: playback,
    isLoading,
    error,
  } = trpc.variants.getPlaybackUrl.useQuery(
    { variantId: asset.id },
    { enabled: Boolean(asset.id) },
  );

  return (
    // biome-ignore lint/a11y/useSemanticElements: real <dialog> needs showModal()/close() + native backdrop styling, out of scope here
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 backdrop-blur-sm"
      onClick={onClose}
      onKeyDown={(e) => {
        if (e.key === "Escape") onClose();
      }}
    >
      <div
        className="relative w-full max-w-3xl overflow-hidden rounded-2xl border border-white/10 bg-black/90"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-3 top-3 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-black/60 text-white/60 backdrop-blur-sm transition-colors hover:bg-black/80 hover:text-white"
        >
          ✕
        </button>

        {isLoading ? (
          <div className="flex aspect-video items-center justify-center bg-white/[0.03]">
            <p className="text-sm text-white/40">Loading playback…</p>
          </div>
        ) : error || !playback ? (
          <div className="flex aspect-video items-center justify-center bg-white/[0.03]">
            <p className="text-sm text-white/40">{error?.message ?? "No playback available"}</p>
          </div>
        ) : (
          <MuxPlayer
            playbackId={playback.playbackId}
            tokens={{ playback: playback.token }}
            metadata={{ video_title: asset.angle || "Asset preview", viewer_user_id: asset.id }}
            streamType="on-demand"
            autoPlay
            style={{ width: "100%", aspectRatio: "16/9" }}
          />
        )}

        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3">
            <p className="text-sm font-medium text-white">{asset.angle || "Untitled"}</p>
            <Badge variant={statusVariant(asset.status)} className="capitalize">
              {asset.status}
            </Badge>
          </div>
          {asset.duration_secs != null && (
            <p className="text-sm text-white/40">{formatDuration(asset.duration_secs)}</p>
          )}
        </div>
      </div>
    </div>
  );
}

function PlayIcon({ className, filled }: { className?: string; filled?: boolean }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      viewBox="0 0 24 24"
      fill={filled ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth={filled ? 0 : 1.5}
      aria-hidden="true"
    >
      {filled ? (
        <path d="M8 5v14l11-7z" />
      ) : (
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M15 10l4.553-2.069A1 1 0 0121 8.845v6.31a1 1 0 01-1.447.894L15 14M3 8a2 2 0 012-2h10a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V8z"
        />
      )}
    </svg>
  );
}
