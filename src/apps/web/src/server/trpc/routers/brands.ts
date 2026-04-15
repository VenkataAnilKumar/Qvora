import { createTRPCRouter, workspaceProcedure } from "../init";

export const brandsRouter = createTRPCRouter({
  list: workspaceProcedure.query(async ({ ctx }) => {
    return {
      orgId: ctx.orgId,
      brands: [],
    };
  }),
});
