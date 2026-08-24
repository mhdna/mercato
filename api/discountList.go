package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createDiscountListRequest struct {
	Name      string    `json:"name" binding:"required"`
	IsActive  bool      `json:"is_active"`
	IsDefault bool      `json:"is_default"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
}

func (server *Server) createDiscountList(ctx *gin.Context) {
	var req createDiscountListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateDiscountListParams{
		Name:      req.Name,
		IsActive:  req.IsActive,
		IsDefault: req.IsDefault,
		ValidFrom: req.ValidFrom,
		ValidTo:   req.ValidTo,
	}

	discountList, err := server.store.CreateDiscountList(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, discountList)
}

type getDiscountListRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getDiscountList(ctx *gin.Context) {
	var req getDiscountListRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	discountList, err := server.store.GetDiscountList(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, discountList)
}

type listDiscountListsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listDiscountLists(ctx *gin.Context) {
	var req listDiscountListsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListDiscountListsParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	discountLists, err := server.store.ListDiscountLists(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountDiscountLists(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"discount_lists": discountLists, "total": total})
}

type updateDiscountListRequest struct {
	ID        int64     `json:"id" binding:"required,min=1"`
	Name      string    `json:"name" binding:"required"`
	IsActive  bool      `json:"is_active"`
	IsDefault bool      `json:"is_default"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
}

func (server *Server) updateDiscountList(ctx *gin.Context) {
	var req updateDiscountListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.UpdateDiscountList(ctx, db.UpdateDiscountListParams{
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
	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}

type createDiscountListItemRequest struct {
	DiscountListID int64 `json:"discount_list_id" binding:"required"`
	ProductID      int64 `json:"product_id" binding:"required"`
	Discount       int16 `json:"discount" binding:"required"`
}

func (server *Server) createDiscountListItem(ctx *gin.Context) {
	var req createDiscountListItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateDiscountListItemParams{
		DiscountListID: req.DiscountListID,
		ProductID:      req.ProductID,
		Discount:       req.Discount,
	}

	item, err := server.store.CreateDiscountListItem(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, item)
}

type listDiscountListItemsRequest struct {
	DiscountListID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) listDiscountListItems(ctx *gin.Context) {
	var req listDiscountListItemsRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	items, err := server.store.ListDiscountListItems(ctx, req.DiscountListID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, items)
}

type deleteDiscountListItemRequest struct {
	DiscountListID int64 `uri:"id" binding:"required,min=1"`
	ProductID      int64 `uri:"product_id" binding:"required,min=1"`
}

func (server *Server) deleteDiscountListItem(ctx *gin.Context) {
	var req deleteDiscountListItemRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteDiscountListItem(ctx, db.DeleteDiscountListItemParams{
		DiscountListID: req.DiscountListID,
		ProductID:      req.ProductID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
