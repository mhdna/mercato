package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createPriceListRequest struct {
	Name      string    `json:"name" binding:"required"`
	IsActive  bool      `json:"is_active"`
	IsDefault bool      `json:"is_default"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
}

func (server *Server) createPriceList(ctx *gin.Context) {
	var req createPriceListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreatePriceListParams{
		Name:      req.Name,
		IsActive:  req.IsActive,
		IsDefault: req.IsDefault,
		ValidFrom: req.ValidFrom,
		ValidTo:   req.ValidTo,
	}

	priceList, err := server.store.CreatePriceList(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Only one price list can be the global default -- demote any other.
	if priceList.IsDefault {
		if err := server.store.UnsetDefaultPriceList(ctx, priceList.ID); err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, priceList)
}

type updatePriceListRequest struct {
	ID        int64     `json:"id" binding:"required,min=1"`
	Name      string    `json:"name" binding:"required"`
	IsActive  bool      `json:"is_active"`
	IsDefault bool      `json:"is_default"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
}

func (server *Server) updatePriceList(ctx *gin.Context) {
	var req updatePriceListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.UpdatePriceList(ctx, db.UpdatePriceListParams{
		ID:        req.ID,
		Name:      req.Name,
		IsActive:  req.IsActive,
		IsDefault: req.IsDefault,
		ValidFrom: req.ValidFrom,
		ValidTo:   req.ValidTo,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if req.IsDefault {
		if err := server.store.UnsetDefaultPriceList(ctx, req.ID); err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}

type deletePriceListRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deletePriceList(ctx *gin.Context) {
	var req deletePriceListRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeletePriceList(ctx, req.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

type getPriceListRequest struct {
	ID int64 `uri:"id" binding:"required,min=0"`
}

func (server *Server) getPriceList(ctx *gin.Context) {
	var req getPriceListRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	priceList, err := server.store.GetPriceList(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, priceList)
}

func (server *Server) listPriceLists(ctx *gin.Context) {
	q, ok := server.bindListPageQuery(ctx)
	if !ok {
		return
	}

	respondList(server, ctx, "price_lists",
		func() ([]db.PriceList, error) {
			return server.store.ListPriceLists(ctx, db.ListPriceListsParams{
				Search: q.Search, PageSize: q.PageSize, PageOffset: q.PageID,
			})
		},
		func() (int64, error) { return server.store.CountPriceListsFiltered(ctx, q.Search) },
	)
}

type createPriceListItemRequest struct {
	PriceListID int64 `json:"price_list_id" binding:"required"`
	ProductID   int64 `json:"product_id" binding:"required"`
	// 0 is a legitimate price (a giveaway line), so no "required" here.
	Price int64 `json:"price" binding:"min=0"`
}

func (server *Server) createPriceListItem(ctx *gin.Context) {
	var req createPriceListItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreatePriceListItemParams{
		PriceListID: req.PriceListID,
		ProductID:   req.ProductID,
		Price:       req.Price,
	}

	item, err := server.store.CreatePriceListItem(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, item)
}

type listPriceListItemsRequest struct {
	PriceListID int64 `uri:"id" binding:"required,min=0"`
}

func (server *Server) listPriceListItems(ctx *gin.Context) {
	var req listPriceListItemsRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	q, ok := server.bindListPageQuery(ctx)
	if !ok {
		return
	}

	respondList(server, ctx, "items",
		func() ([]db.ListPriceListItemsWithProductRow, error) {
			return server.store.ListPriceListItemsWithProduct(ctx, db.ListPriceListItemsWithProductParams{
				PriceListID: req.PriceListID,
				Search:      q.Search,
				PageSize:    q.PageSize,
				PageOffset:  q.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountPriceListItemsWithProduct(ctx, db.CountPriceListItemsWithProductParams{
				PriceListID: req.PriceListID,
				Search:      q.Search,
			})
		},
	)
}

type deletePriceListItemRequest struct {
	PriceListID int64 `uri:"id" binding:"required,min=0"`
	ProductID   int64 `uri:"product_id" binding:"required,min=0"`
}

func (server *Server) deletePriceListItem(ctx *gin.Context) {
	var req deletePriceListItemRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeletePriceListItem(ctx, db.DeletePriceListItemParams{
		PriceListID: req.PriceListID,
		ProductID:   req.ProductID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ---- branch assignments -------------------------------------------------
//
// A branch can have at most one price list enabled at a time (branch_id is
// the PK of price_list_branches). Assigning a branch that already belongs to
// a different list is rejected rather than silently reassigned, so the
// operator has to consciously move it.

type priceListBranchesURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) listPriceListBranches(ctx *gin.Context) {
	var req priceListBranchesURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	branchIDs, err := server.store.ListPriceListBranchIDs(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"branch_ids": branchIDs})
}

type setPriceListBranchesRequest struct {
	BranchIDs []int64 `json:"branch_ids"`
}

func (server *Server) setPriceListBranches(ctx *gin.Context) {
	var uri priceListBranchesURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req setPriceListBranchesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	// Reject the whole request if any branch is already claimed by another
	// list -- checked before any write so a conflict leaves things untouched.
	for _, branchID := range req.BranchIDs {
		existing, err := server.store.GetPriceListBranch(ctx, branchID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		if existing.PriceListID != uri.ID {
			server.writeError(ctx, http.StatusConflict, fmt.Errorf(
				"branch %d is already assigned to price list %d; remove it there first",
				branchID, existing.PriceListID))
			return
		}
	}

	if err := server.store.DeletePriceListBranchesForList(ctx, uri.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	for _, branchID := range req.BranchIDs {
		if err := server.store.UpsertPriceListBranch(ctx, db.UpsertPriceListBranchParams{
			BranchID:    branchID,
			PriceListID: uri.ID,
		}); err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"branch_ids": req.BranchIDs})
}
