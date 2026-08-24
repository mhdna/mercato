package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchSettingsRequest carries only the "business-identity" subset of
// kashi-pos's local settings table -- everything a branch operator might
// need to see/edit remotely. Hardware/device fields (printer, screen port,
// theme, counter number) and the branch's own credentials (api_key) never
// leave the branch; kashi has no business modeling them centrally.
type branchSettingsRequest struct {
	BranchName          string          `json:"branch_name"`
	TaxRate             float64         `json:"tax_rate"`
	RoundingMode        string          `json:"rounding_mode"`
	RoundingCurrency    string          `json:"rounding_currency"`
	ExchangeRate        int64           `json:"exchange_rate"`
	ExchangeWindowHours int64           `json:"exchange_window_hours"`
	MarketName          string          `json:"market_name"`
	MarketPhone         string          `json:"market_phone"`
	MarketDescription   string          `json:"market_description"`
	ReturnPolicy        string          `json:"return_policy"`
	Website             string          `json:"website"`
	Instagram           string          `json:"instagram"`
	SocialPlatforms     json.RawMessage `json:"social_platforms"`
	SocialHandles       json.RawMessage `json:"social_handles"`
}

// putBranchSettings is how a branch reports its current settings snapshot
// -- called on every sync tick by kashi-pos (see runSyncTick/pushSettings
// in the kashi-pos repo), not just on change, since it's a plain upsert of
// the whole row rather than an append-only event. Idempotent by
// construction: pushing the same snapshot twice is a no-op in effect.
func (server *Server) putBranchSettings(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	socialPlatforms := req.SocialPlatforms
	if len(socialPlatforms) == 0 {
		socialPlatforms = json.RawMessage("[]")
	}
	socialHandles := req.SocialHandles
	if len(socialHandles) == 0 {
		socialHandles = json.RawMessage("{}")
	}

	settings, err := server.store.UpsertBranchSettings(ctx, db.UpsertBranchSettingsParams{
		BranchID:            branchID,
		BranchName:          req.BranchName,
		TaxRate:             req.TaxRate,
		RoundingMode:        req.RoundingMode,
		RoundingCurrency:    req.RoundingCurrency,
		ExchangeRate:        req.ExchangeRate,
		ExchangeWindowHours: req.ExchangeWindowHours,
		MarketName:          req.MarketName,
		MarketPhone:         req.MarketPhone,
		MarketDescription:   req.MarketDescription,
		ReturnPolicy:        req.ReturnPolicy,
		Website:             req.Website,
		Instagram:           req.Instagram,
		SocialPlatforms:     socialPlatforms,
		SocialHandles:       socialHandles,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"settings": settings})
}

// getBranchSettings is the admin-facing read side, used by kashi's UI to
// show a branch's current settings snapshot.
func (server *Server) getBranchSettings(ctx *gin.Context) {
	var req branchIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	settings, err := server.store.GetBranchSettings(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"settings": settings})
}
