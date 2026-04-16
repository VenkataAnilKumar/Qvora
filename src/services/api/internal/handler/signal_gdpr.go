package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/domain/signal"
)

func HandleSignalGDPRCleanup(c echo.Context) error { return signal.HandleSignalGDPRCleanup(c) }
