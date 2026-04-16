package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/media"
)

func SubmitJob(c echo.Context) error       { return media.SubmitJob(c) }
func GetJob(c echo.Context) error          { return media.GetJob(c) }
func UpdateJobStatus(c echo.Context) error { return media.UpdateJobStatus(c) }
func ListJobs(c echo.Context) error        { return media.ListJobs(c) }
func StreamJob(c echo.Context) error       { return media.StreamJob(c) }
