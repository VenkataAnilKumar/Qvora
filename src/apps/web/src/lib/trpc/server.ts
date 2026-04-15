import "server-only";
import { createCallerFactory, createTRPCContext } from "@/server/trpc/init";
import { appRouter } from "@/server/trpc/routers";
import { headers } from "next/headers";
import { cache } from "react";

// Server-side tRPC caller — for use in Server Components and Route Handlers
const createCaller = createCallerFactory(appRouter);

export const createServerClient = cache(async () => {
  const heads = await headers();
  return createCaller(
    await createTRPCContext({
      headers: heads,
    }),
  );
});
