package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/asset"
)

func ListAssets(c echo.Context) error  { return asset.ListAssets(c) }
func DeleteAsset(c echo.Context) error { return asset.DeleteAsset(c) }
