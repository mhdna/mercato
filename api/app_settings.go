package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// getAppSettings returns the single global settings row (seeded by
// migration 000050, always present -- there is exactly one, id = 1).
func (server *Server) getAppSettings(ctx *gin.Context) {
	settings, err := server.store.GetAppSettings(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"settings": settings})
}

var validActivityDisplayModes = map[string]bool{"notification": true, "appbar": true}

type updateAppSettingsRequest struct {
	FinancialsHighSeasonMonths []int   `json:"financials_high_season_months"`
	ActivityMessageSeconds     int16   `json:"activity_message_seconds"`
	ActivityDisplayMode        string  `json:"activity_display_mode"`
	BarcodeLabelWidth          float32 `json:"barcode_label_width"`
	BarcodeLabelHeight         float32 `json:"barcode_label_height"`
}

// updateAppSettings replaces the whole global settings row -- the client
// always sends the full object (see ui/src/stores/settings.js), so this is
// a plain overwrite rather than a partial patch.
func (server *Server) updateAppSettings(ctx *gin.Context) {
	var req updateAppSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	for _, month := range req.FinancialsHighSeasonMonths {
		if month < 1 || month > 12 {
			server.writeError(ctx, http.StatusBadRequest, errors.New("financials_high_season_months must contain months between 1 and 12"))
			return
		}
	}
	if req.ActivityMessageSeconds < 1 {
		server.writeError(ctx, http.StatusBadRequest, errors.New("activity_message_seconds must be greater than 0"))
		return
	}
	if !validActivityDisplayModes[req.ActivityDisplayMode] {
		server.writeError(ctx, http.StatusBadRequest, errors.New("activity_display_mode must be 'notification' or 'appbar'"))
		return
	}
	if req.BarcodeLabelWidth <= 0 || req.BarcodeLabelHeight <= 0 {
		server.writeError(ctx, http.StatusBadRequest, errors.New("barcode label dimensions must be greater than 0"))
		return
	}

	months, err := json.Marshal(req.FinancialsHighSeasonMonths)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	settings, err := server.store.UpdateAppSettings(ctx, db.UpdateAppSettingsParams{
		FinancialsHighSeasonMonths: months,
		ActivityMessageSeconds:     req.ActivityMessageSeconds,
		ActivityDisplayMode:        req.ActivityDisplayMode,
		BarcodeLabelWidth:          req.BarcodeLabelWidth,
		BarcodeLabelHeight:         req.BarcodeLabelHeight,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"settings": settings})
}
