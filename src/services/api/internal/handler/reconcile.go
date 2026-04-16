package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/media"
)

func ReconcileStuckJobs(c echo.Context) error { return media.ReconcileStuckJobs(c) }
