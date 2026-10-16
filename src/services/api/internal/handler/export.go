package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/domain/asset"
)

func ListExports(c echo.Context) error    { return asset.ListExports(c) }
func CreateExports(c echo.Context) error  { return asset.CreateExports(c) }
func DownloadExport(c echo.Context) error { return asset.DownloadExport(c) }
