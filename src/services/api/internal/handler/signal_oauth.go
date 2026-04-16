package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/domain/signal"
)

func InitiateSignalOAuth(c echo.Context) error       { return signal.InitiateSignalOAuth(c) }
func HandleSignalOAuthCallback(c echo.Context) error { return signal.HandleSignalOAuthCallback(c) }
