package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/asset"
)

func ListExports(c echo.Context) error { return asset.ListExports(c) }
