import { Langfuse } from "langfuse";

type BriefTraceInput = {
  traceName: string;
  userId: string;
  orgId: string;
  briefId?: string;
  model: string;
  input: unknown;
  output: unknown;
  metadata?: Record<string, unknown>;
};

let singleton: Langfuse | null = null;

function getLangfuseClient(): Langfuse | null {
  const publicKey = process.env.LANGFUSE_PUBLIC_KEY;
  const secretKey = process.env.LANGFUSE_SECRET_KEY;

  if (!publicKey || !secretKey) {
    return null;
  }

  if (singleton) {
    return singleton;
  }

  singleton = new Langfuse({
    publicKey,
    secretKey,
    baseUrl: process.env.LANGFUSE_BASE_URL,
    environment: process.env.NODE_ENV,
  });

  return singleton;
}

export async function recordBriefTrace(payload: BriefTraceInput): Promise<void> {
  const client = getLangfuseClient();
  if (!client) {
    return;
  }

  try {
    const trace = client.trace({
      name: payload.traceName,
      userId: payload.userId,
      sessionId: payload.orgId,
      metadata: {
        orgId: payload.orgId,
        briefId: payload.briefId,
        ...payload.metadata,
      },
      tags: ["brief", "phase2"],
    });

    const generation = trace.generation({
      name: payload.traceName,
      model: payload.model,
      input: payload.input,
      output: payload.output,
      metadata: payload.metadata,
    });

    generation.end({ output: payload.output });
    await client.flushAsync();
  } catch {
    // Tracing must never break brief generation/edit flows.
  }
}
