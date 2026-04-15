package task

// Task type constants — used as asynq queue task type identifiers
const (
	TypeScrape                       = "job:scrape"
	TypeGenerate                     = "job:generate"
	TypePostprocess                  = "job:postprocess"
	TypeSignalRecommendationsRefresh = "signal:recommendations:refresh"
	TypeSignalMetricsSync            = "signal:metrics:sync"
	TypeSignalGDPRCleanup            = "signal:gdpr:cleanup"
	TypeJobReconcileStuck            = "jobs:reconcile:stuck"
)
