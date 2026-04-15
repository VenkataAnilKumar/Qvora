import { createTRPCRouter, workspaceProcedure } from "../init";

export const projectsRouter = createTRPCRouter({
  list: workspaceProcedure.query(async ({ ctx }) => {
    return {
      orgId: ctx.orgId,
      projects: [],
    };
  }),
});
