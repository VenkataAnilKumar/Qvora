"use client";

import { useEffect, useRef, useState } from "react";
import { Badge, Spinner } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

// Raw SSE event shape emitted by the Go API (flat, snake_case)
type RawSSEEvent = {
  type: string;
  job_id: string;
  status: string;
  message: string;
  progress: number;
};

const TERMINAL = new Set(["complete", "failed"]);

const PROGRESS_MAP: Record<string, number> = {
  queued: 5,
  scraping: 20,
  briefing: 55,
  generating: 75,
  postprocessing: 90,
  complete: 100,
  failed: 100,
};

const PROGRESS_WIDTH_CLASS: Record<string, string> = {
  queued: "w-[5%]",
  scraping: "w-[20%]",
  briefing: "w-[55%]",
  generating: "w-[75%]",
  postprocessing: "w-[90%]",
  complete: "w-full",
  failed: "w-full",
};

const LABEL_MAP: Record<string, string> = {
  queued: "Job accepted — entering pipeline",
  scraping: "Scraping product page via Modal Playwright",
  briefing: "Generating creative brief with GPT-4o",
  generating: "Generating video variants via FAL.AI",
  postprocessing: "Postprocessing & encoding via Rust",
  complete: "Generation complete",
  failed: "Job failed",
};

interface JobProgressProps {
  jobId: string;
  productUrl: string;
}

export function JobProgress({ jobId, productUrl }: JobProgressProps) {
  const [sseStatus, setSseStatus] = useState<string>("queued");
  const [sseProgress, setSseProgress] = useState<number>(5);
  const [shouldPoll, setShouldPoll] = useState(true);
  const esRef = useRef<EventSource | null>(null);

  // ── SSE stream for real-time updates ─────────────────────────────────────
  useEffect(() => {
    if (!shouldPoll) return;

    const es = new EventSource(`/api/generation/${jobId}/stream`);
    esRef.current = es;

    es.onmessage = (e: MessageEvent<string>) => {
      try {
        const event = JSON.parse(e.data) as RawSSEEvent;
        setSseStatus(event.status);
        setSseProgress(event.progress);
        if (TERMINAL.has(event.status)) {
          setShouldPoll(false);
          es.close();
        }
      } catch {
        // non-fatal parse error — polling fallback will catch the state
      }
    };

    es.onerror = () => {
      es.close(); // polling fallback takes over
    };

    return () => {
      es.close();
    };
  }, [jobId, shouldPoll]);

  // ── Polling fallback — drives final state when SSE closes ─────────────────
  const { data: job } = trpc.generation.getJob.useQuery(
    { jobId },
    { refetchInterval: shouldPoll ? 2000 : false },
  );

  // Polling result wins once it disagrees with (stale) SSE snapshot
  useEffect(() => {
    if (job?.status) {
      setSseStatus(job.status);
      setSseProgress((current) => PROGRESS_MAP[job.status] ?? current);
      if (TERMINAL.has(job.status)) setShouldPoll(false);
    }
  }, [job?.status]); // eslint-disable-line react-hooks/exhaustive-deps

  const status = sseStatus;
  const progress = sseProgress;
  const progressWidthClass = PROGRESS_WIDTH_CLASS[status] ?? "w-[5%]";
  const isDone = status === "complete";
  const isFailed = status === "failed";
  const isActive = !isDone && !isFailed;

  return (
    <div className="space-y-4 rounded-2xl border border-white/8 bg-black/20 p-5">
      {/* Header row */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-white/90">{productUrl}</p>
          <p className="mt-0.5 font-mono text-xs text-white/38">{jobId}</p>
        </div>
        {isDone ? (
          <Badge variant="success">Complete</Badge>
        ) : isFailed ? (
          <Badge variant="destructive">Failed</Badge>
        ) : (
          <div className="flex shrink-0 items-center gap-2">
            <Spinner size="sm" />
            <Badge variant="outline">{status}</Badge>
          </div>
        )}
      </div>

      {/* Progress bar */}
      <div className="space-y-1.5">
        <div className="flex justify-between text-xs text-white/45">
          <span>{LABEL_MAP[status] ?? status}</span>
          <span>{progress}%</span>
        </div>
        <div className="h-1.5 w-full overflow-hidden rounded-full bg-white/8">
          <div
            className={[
              "h-full rounded-full transition-all duration-700",
              progressWidthClass,
              isDone
                ? "bg-[var(--color-convert-green)]"
                : isFailed
                  ? "bg-[var(--color-signal-red)]"
                  : "bg-[var(--color-volt)]",
            ].join(" ")}
          />
        </div>
      </div>

      {/* Terminal action row */}
      {isDone && (
        <p className="text-xs text-[var(--color-convert-green)]">
          Brief ready — variants available in the generation run.
        </p>
      )}
      {isFailed && (
        <p className="text-xs text-[var(--color-signal-red)]">
          Check worker logs. Verify MODAL_SCRAPER_ENDPOINT and OPENAI_API_KEY are set.
        </p>
      )}
      {isActive && (
        <p className="text-xs text-white/35">Status updates every ~1 second via SSE stream.</p>
      )}
    </div>
  );
}
