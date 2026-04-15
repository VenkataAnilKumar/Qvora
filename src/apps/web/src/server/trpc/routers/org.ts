import { createTRPCRouter, workspaceProcedure } from "../init";

export const orgRouter = createTRPCRouter({
  current: workspaceProcedure.query(async ({ ctx }) => {
    return {
      orgId: ctx.orgId,
      role: ctx.orgRole ?? "member",
    };
  }),
});
