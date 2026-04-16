package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/asset"
)

func GetVariantPlaybackURL(c echo.Context) error  { return asset.GetVariantPlaybackURL(c) }
func UpdateVariantFalRequest(c echo.Context) error { return asset.UpdateVariantFalRequest(c) }
