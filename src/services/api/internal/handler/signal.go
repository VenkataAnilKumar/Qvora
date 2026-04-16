package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/domain/signal"
)

func ListSignalConnections(c echo.Context) error       { return signal.ListSignalConnections(c) }
func UpsertSignalConnection(c echo.Context) error      { return signal.UpsertSignalConnection(c) }
func PatchSignalConnectionHealth(c echo.Context) error { return signal.PatchSignalConnectionHealth(c) }
func UpsertSignalMetrics(c echo.Context) error         { return signal.UpsertSignalMetrics(c) }
func GetSignalDashboard(c echo.Context) error          { return signal.GetSignalDashboard(c) }
func DetectSignalFatigue(c echo.Context) error         { return signal.DetectSignalFatigue(c) }
func ListSignalFatigueEvents(c echo.Context) error     { return signal.ListSignalFatigueEvents(c) }
func GetSignalRecommendations(c echo.Context) error    { return signal.GetSignalRecommendations(c) }
func RefreshAllSignalRecommendations(c echo.Context) error {
	return signal.RefreshAllSignalRecommendations(c)
}
func CreateSignalRecommendationFeedback(c echo.Context) error {
	return signal.CreateSignalRecommendationFeedback(c)
}
func ListSignalRecommendationFeedbackByAngle(c echo.Context) error {
	return signal.ListSignalRecommendationFeedbackByAngle(c)
}
