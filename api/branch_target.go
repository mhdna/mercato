package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// branchTargetColorPalette is cycled through (by how many targets a branch
// already has) to assign a color when the admin doesn't specify one --
// distinct enough from each other to tell layered bars apart at a glance,
// without needing a color picker to be mandatory.
var branchTargetColorPalette = []string{
	"#4C6EF5", "#12B886", "#F59F00", "#E64980",
	"#7048E8", "#15AABF", "#FA5252", "#82C91E",
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

var errInvalidBranchTargetColor = fmt.Errorf("color must be a hex value like #4C6EF5")

func assignBranchTargetColor(ctx context.Context, store db.Store, branchID int64) (string, error) {
	existing, err := store.ListBranchTargetsForBranch(ctx, branchID)
	if err != nil {
		return "", err
	}
	return branchTargetColorPalette[len(existing)%len(branchTargetColorPalette)], nil
}

// branchTargetResponse adds the live-computed `achieved` amount to a target
// row -- progress is never stored, only derived from branch_invoices at
// read time, so it can't drift from actual sales.
type branchTargetResponse struct {
	db.BranchTarget
	Achieved int64 `json:"achieved"`
}

func (server *Server) targetWithProgress(ctx *gin.Context, target db.BranchTarget) (branchTargetResponse, error) {
	achieved, err := server.store.SumBranchRevenueForRange(ctx, db.SumBranchRevenueForRangeParams{
		BranchID: target.BranchID,
		DateFrom: target.DateFrom,
		DateTo:   target.DateTo,
	})
	if err != nil {
		return branchTargetResponse{}, err
	}
	return branchTargetResponse{BranchTarget: target, Achieved: achieved}, nil
}

type createBranchTargetRequest struct {
	DateFrom     time.Time `json:"date_from" binding:"required"`
	DateTo       time.Time `json:"date_to" binding:"required"`
	TargetAmount int64     `json:"target_amount" binding:"required,min=1"`
	// Optional hex color (e.g. "#4C6EF5"); left blank, one is auto-assigned
	// from branchTargetColorPalette.
	Color string `json:"color"`
}

func (server *Server) createBranchTarget(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var req createBranchTargetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.Color != "" && !hexColorPattern.MatchString(req.Color) {
		server.writeError(ctx, http.StatusBadRequest, errInvalidBranchTargetColor)
		return
	}

	color := req.Color
	if color == "" {
		assigned, err := assignBranchTargetColor(ctx, server.store, uri.ID)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		color = assigned
	}

	target, err := server.store.CreateBranchTarget(ctx, db.CreateBranchTargetParams{
		BranchID:     uri.ID,
		DateFrom:     req.DateFrom,
		DateTo:       req.DateTo,
		TargetAmount: req.TargetAmount,
		Color:        color,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	response, err := server.targetWithProgress(ctx, target)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// listBranchTargets returns every target for a branch, low-to-high by
// amount (see ListBranchTargetsForBranch), each with its current progress
// -- the admin UI renders these as layered bars in that same order.
func (server *Server) listBranchTargets(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	targets, err := server.store.ListBranchTargetsForBranch(ctx, uri.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	responses := make([]branchTargetResponse, 0, len(targets))
	for _, target := range targets {
		response, err := server.targetWithProgress(ctx, target)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		responses = append(responses, response)
	}
	ctx.JSON(http.StatusOK, envelope{"targets": responses})
}

type updateBranchTargetRequest struct {
	ID           int64     `uri:"id" binding:"required,min=1"`
	DateFrom     time.Time `json:"date_from" binding:"required"`
	DateTo       time.Time `json:"date_to" binding:"required"`
	TargetAmount int64     `json:"target_amount" binding:"required,min=1"`
	// Optional; blank keeps the target's current color rather than clearing it.
	Color string `json:"color"`
}

func (server *Server) updateBranchTarget(ctx *gin.Context) {
	var req updateBranchTargetRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.Color != "" && !hexColorPattern.MatchString(req.Color) {
		server.writeError(ctx, http.StatusBadRequest, errInvalidBranchTargetColor)
		return
	}

	color := req.Color
	if color == "" {
		existing, err := server.store.GetBranchTarget(ctx, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				server.writeError(ctx, http.StatusNotFound, err)
				return
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		color = existing.Color
	}

	target, err := server.store.UpdateBranchTarget(ctx, db.UpdateBranchTargetParams{
		ID:           req.ID,
		DateFrom:     req.DateFrom,
		DateTo:       req.DateTo,
		TargetAmount: req.TargetAmount,
		Color:        color,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	response, err := server.targetWithProgress(ctx, target)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

type deleteBranchTargetRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// deleteBranchTarget flips is_active rather than removing the row -- see
// 000046_branch_target_soft_delete.up.sql. Kashi's own UI (which reads
// ListBranchTargetsForBranch) already filters inactive targets out, so
// this still looks and behaves like a delete from the admin's side; the
// row survives so kashi-pos's sync-since pull can see the change and
// remove its local copy.
func (server *Server) deleteBranchTarget(ctx *gin.Context) {
	var req deleteBranchTargetRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.SetBranchTargetActive(ctx, db.SetBranchTargetActiveParams{
		ID:       req.ID,
		IsActive: false,
	}); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
