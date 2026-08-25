package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// listSalespersons returns every salesperson on a branch's roster,
// including inactive ones -- the admin UI needs to see and reactivate a
// deactivated salesperson, not just the active subset kashi-pos syncs down.
func (server *Server) listSalespersons(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	salespersons, err := server.store.ListSalespersonsForBranch(ctx, sql.NullInt64{Int64: uri.ID, Valid: true})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"salespersons": salespersons})
}

type createSalespersonRequest struct {
	Name string `json:"name" binding:"required"`
}

func (server *Server) createSalesperson(ctx *gin.Context) {
	var uri branchIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var req createSalespersonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	salesperson, err := server.store.CreateBranchSalesperson(ctx, db.CreateBranchSalespersonParams{
		Name:     req.Name,
		BranchID: sql.NullInt64{Int64: uri.ID, Valid: true},
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.notify(uri.ID, branchWSMessage{Type: "salesperson_updated"})
	ctx.JSON(http.StatusOK, salesperson)
}

type updateSalespersonRequest struct {
	ID   int64  `uri:"id" binding:"required,min=1"`
	Name string `json:"name" binding:"required"`
}

func (server *Server) updateSalesperson(ctx *gin.Context) {
	var req updateSalespersonRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	existing, err := server.store.GetSalesperson(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	salesperson, err := server.store.UpdateBranchSalespersonName(ctx, db.UpdateBranchSalespersonNameParams{
		ID:       req.ID,
		Name:     req.Name,
		BranchID: existing.BranchID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if existing.BranchID.Valid {
		server.branchHub.notify(existing.BranchID.Int64, branchWSMessage{Type: "salesperson_updated"})
	}
	ctx.JSON(http.StatusOK, salesperson)
}

type setSalespersonActiveRequest struct {
	ID       int64 `uri:"id" binding:"required,min=1"`
	IsActive bool  `json:"is_active"`
}

func (server *Server) setSalespersonActive(ctx *gin.Context) {
	var req setSalespersonActiveRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	salesperson, err := server.store.SetSalespersonActive(ctx, db.SetSalespersonActiveParams{
		ID:       req.ID,
		IsActive: req.IsActive,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if salesperson.BranchID.Valid {
		server.branchHub.notify(salesperson.BranchID.Int64, branchWSMessage{Type: "salesperson_updated"})
	}
	ctx.JSON(http.StatusOK, salesperson)
}
