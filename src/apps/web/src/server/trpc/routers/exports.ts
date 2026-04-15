import { createTRPCRouter, workspaceProcedure } from "../init";

export const exportsRouter = createTRPCRouter({
  list: workspaceProcedure.query(async ({ ctx }) => {
    return {
      orgId: ctx.orgId,
      exports: [],
    };
  }),
});
