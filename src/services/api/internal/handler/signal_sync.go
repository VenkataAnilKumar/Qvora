package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/domain/signal"
)

func SyncSignalMetricsAll(c echo.Context) error { return signal.SyncSignalMetricsAll(c) }
