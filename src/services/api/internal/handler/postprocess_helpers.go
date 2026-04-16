package handler

// isIdempotentCallback returns true when rowsAffected == 0, meaning the
// callback record already exists and the request can be safely no-op'd.
func isIdempotentCallback(rowsAffected int64) bool {
	return rowsAffected == 0
}

// targetVariantStatus resolves the next variant status given a postprocess
// callback outcome and the variant's current status.
// Terminal statuses ("complete") are never overwritten.
func targetVariantStatus(callbackStatus, currentStatus string) string {
	if currentStatus == "complete" {
		return "complete"
	}
	if callbackStatus == "success" {
		return "postprocessing"
	}
	return "failed"
}

// targetJobStatus resolves the next job status given a postprocess callback
// outcome and the job's current status.
// Terminal statuses ("complete") are never overwritten.
func targetJobStatus(callbackStatus, currentStatus string) string {
	if currentStatus == "complete" {
		return "complete"
	}
	if callbackStatus == "success" {
		return "postprocessing"
	}
	return "failed"
}
