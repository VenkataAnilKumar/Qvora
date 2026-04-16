package handler

import (
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/domain/media"
)

func MuxWebhook(c echo.Context) error   { return media.MuxWebhook(c) }
func FalWebhook(c echo.Context) error   { return media.FalWebhook(c) }
func ClerkWebhook(c echo.Context) error { return media.ClerkWebhook(c) }
