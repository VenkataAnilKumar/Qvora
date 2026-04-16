package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/brief"
)

func CreateBrief(c echo.Context) error         { return brief.CreateBrief(c) }
func ListBriefs(c echo.Context) error          { return brief.ListBriefs(c) }
func GetBrief(c echo.Context) error            { return brief.GetBrief(c) }
func UpdateBriefContent(c echo.Context) error  { return brief.UpdateBriefContent(c) }
