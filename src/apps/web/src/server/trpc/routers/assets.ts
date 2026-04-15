import { createTRPCRouter, workspaceProcedure } from "../init";

export const assetsRouter = createTRPCRouter({
  list: workspaceProcedure.query(async ({ ctx }) => {
    return {
      orgId: ctx.orgId,
      assets: [],
    };
  }),
});
