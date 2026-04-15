import { createTRPCRouter } from "../init";
import { assetsRouter } from "./assets";
import { brandsRouter } from "./brands";
import { briefsRouter } from "./briefs";
import { exportsRouter } from "./exports";
import { generationRouter } from "./generation";
import { jobsRouter } from "./jobs";
import { orgRouter } from "./org";
import { projectsRouter } from "./projects";
import { systemRouter } from "./system";
import { workspaceRouter } from "./workspace";

export const appRouter = createTRPCRouter({
  assets: assetsRouter,
  brands: brandsRouter,
  briefs: briefsRouter,
  exports: exportsRouter,
  generation: generationRouter,
  jobs: jobsRouter,
  org: orgRouter,
  projects: projectsRouter,
  system: systemRouter,
  workspace: workspaceRouter,
});

export type AppRouter = typeof appRouter;
