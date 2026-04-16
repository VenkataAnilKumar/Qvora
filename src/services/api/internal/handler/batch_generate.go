package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/brief"
)

func BatchGenerateVariants(c echo.Context) error { return brief.BatchGenerateVariants(c) }
