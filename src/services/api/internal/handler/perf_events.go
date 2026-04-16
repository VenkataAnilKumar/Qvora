package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/media"
)

func HandleCreatePerfEvent(c echo.Context) error { return media.HandleCreatePerfEvent(c) }
func HandleCreateCostEvent(c echo.Context) error { return media.HandleCreateCostEvent(c) }
func HandlePatchAvatarJob(c echo.Context) error  { return media.HandlePatchAvatarJob(c) }
