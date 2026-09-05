package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchSettingsRequest is the per-branch config kashi mirrors centrally so
// an admin can see and remotely edit it. The branch reports the whole set
// on every sync tick (branchSettingsPayload in kashi-pos); an admin edit is
// an update_settings branch command carrying the subset it wants changed.
//
// Still branch-only, never modelled here: the branch's api_key, its theme,
// and its counter/register number.
//
// Fields added after the first version are pointers so a snapshot from an
// older POS build that omits them leaves the stored value alone rather than
// zeroing it. A nil pointer on first insert falls back to the column
// default.
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
	// SearchButtonEnabled is the one original field kashi both reads and
	// writes. Defaults to true when a branch omits it (older POS build).
	SearchButtonEnabled *bool `json:"search_button_enabled"`

	// --- behaviour feature flags (POS: feature_flags rows) ---
	CustomItemDiscountsEnabled    *bool `json:"custom_item_discounts_enabled"`
	CustomItemPricesEnabled       *bool `json:"custom_item_prices_enabled"`
	PerUnitItemPricesEnabled      *bool `json:"per_unit_item_prices_enabled"`
	PriceChangeManualOverrideMode *bool `json:"price_change_manual_override_mode"`
	InvoiceKeyboardMode           *bool `json:"invoice_keyboard_mode"`
	ClientRequired                *bool `json:"client_required"`

	// --- non-admin page-access locks (POS: page_unlock_* flags) ---
	PageUnlockClients      *bool `json:"page_unlock_clients"`
	PageUnlockInventory    *bool `json:"page_unlock_inventory"`
	PageUnlockTransfers    *bool `json:"page_unlock_transfers"`
	PageUnlockAttendance   *bool `json:"page_unlock_attendance"`
	PageUnlockSalespersons *bool `json:"page_unlock_salespersons"`

	// --- printing (POS: settings columns) ---
	PrinterSize      *string  `json:"printer_size"`
	ReceiptWidth     *int64   `json:"receipt_width"`
	ReceiptHeight    *int64   `json:"receipt_height"`
	ReceiptEnabled   *bool    `json:"receipt_enabled"`
	ReceiptFont      *string  `json:"receipt_font"`
	ReceiptBodyFont  *string  `json:"receipt_body_font"`
	ReceiptTitleSize *float64 `json:"receipt_title_size"`
	ReceiptBodySize  *float64 `json:"receipt_body_size"`
	ReceiptCutoff    *bool    `json:"receipt_cutoff"`

	// --- per-machine device identifiers (reported for display; the admin
	// UI defaults these to "managed at the branch") ---
	PrinterID      *string `json:"printer_id"`
	ReceiptPrinter *string `json:"receipt_printer"`
	ScreenPort     *string `json:"screen_port"`

	// --- attendance device config (POS: attendance_config row + one flag) ---
	AkuvoxIP                           *string `json:"akuvox_ip"`
	AkuvoxUsername                     *string `json:"akuvox_username"`
	AttendanceEnabled                  *bool   `json:"attendance_enabled"`
	AttendanceDuplicateIntervalSeconds *int64  `json:"attendance_duplicate_interval_seconds"`
	AttendanceCashierHistory           *bool   `json:"attendance_cashier_history"`
}

func orZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func orTrue(p *bool) bool {
	if p == nil {
		return true
	}
	return *p
}

// putBranchSettings is how a branch reports its current settings snapshot
// -- called on every sync tick by kashi-pos (see runSyncTick/pushSettings
// in the kashi-pos repo), not just on change, since it's a plain upsert of
// the whole row rather than an append-only event. Idempotent by
// construction: pushing the same snapshot twice is a no-op in effect.
//
// managed_locally is deliberately not touched here -- it's admin-owned
// state, set only via putBranchSettingsManagedLocally.
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
		SearchButtonEnabled: orTrue(req.SearchButtonEnabled),

		CustomItemDiscountsEnabled:    orZero(req.CustomItemDiscountsEnabled),
		CustomItemPricesEnabled:       orZero(req.CustomItemPricesEnabled),
		PerUnitItemPricesEnabled:      orZero(req.PerUnitItemPricesEnabled),
		PriceChangeManualOverrideMode: orZero(req.PriceChangeManualOverrideMode),
		InvoiceKeyboardMode:           orZero(req.InvoiceKeyboardMode),
		ClientRequired:                orZero(req.ClientRequired),

		PageUnlockClients:      orZero(req.PageUnlockClients),
		PageUnlockInventory:    orZero(req.PageUnlockInventory),
		PageUnlockTransfers:    orZero(req.PageUnlockTransfers),
		PageUnlockAttendance:   orZero(req.PageUnlockAttendance),
		PageUnlockSalespersons: orZero(req.PageUnlockSalespersons),

		PrinterSize:      orZero(req.PrinterSize),
		ReceiptWidth:     orZero(req.ReceiptWidth),
		ReceiptHeight:    orZero(req.ReceiptHeight),
		ReceiptEnabled:   orZero(req.ReceiptEnabled),
		ReceiptFont:      orZero(req.ReceiptFont),
		ReceiptBodyFont:  orZero(req.ReceiptBodyFont),
		ReceiptTitleSize: orZero(req.ReceiptTitleSize),
		ReceiptBodySize:  orZero(req.ReceiptBodySize),
		ReceiptCutoff:    orZero(req.ReceiptCutoff),

		PrinterID:      orZero(req.PrinterID),
		ReceiptPrinter: orZero(req.ReceiptPrinter),
		ScreenPort:     orZero(req.ScreenPort),

		AkuvoxIp:                           orZero(req.AkuvoxIP),
		AkuvoxUsername:                     orZero(req.AkuvoxUsername),
		AttendanceEnabled:                  orZero(req.AttendanceEnabled),
		AttendanceDuplicateIntervalSeconds: orZero(req.AttendanceDuplicateIntervalSeconds),
		AttendanceCashierHistory:           orZero(req.AttendanceCashierHistory),
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

type putBranchSettingsManagedLocallyRequest struct {
	// The set of group keys the branch owns locally. The admin UI renders
	// those groups read-only and never puts them in an update_settings
	// command payload. Known groups: "secrets", "device_ids".
	ManagedLocally []string `json:"managed_locally"`
}

// putBranchSettingsManagedLocally records which settings groups the admin
// has handed back to the branch. Admin-only; the branch never sets this.
func (server *Server) putBranchSettingsManagedLocally(ctx *gin.Context) {
	var idReq branchIDRequest
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var body putBranchSettingsManagedLocallyRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if body.ManagedLocally == nil {
		body.ManagedLocally = []string{}
	}

	settings, err := server.store.SetBranchSettingsManagedLocally(ctx, db.SetBranchSettingsManagedLocallyParams{
		BranchID:       idReq.ID,
		ManagedLocally: body.ManagedLocally,
	})
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
