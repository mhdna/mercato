package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// Batch-delete endpoints for the management tables. Each is a POST (not
// DELETE) so the id list travels in a JSON body. The underlying query is a
// single `DELETE ... WHERE id = ANY(...)` statement, which is atomic: an
// FK violation on any one row rolls the whole statement back, so the
// response is all-or-nothing and a 409 means nothing was deleted.

var errBulkInUse = errors.New("one or more of the selected rows are still in use and were not deleted")

type bulkDeleteIDsRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1,max=500,dive,min=1"`
}

func (server *Server) bulkRespond(ctx *gin.Context, rows int64, err error) {
	if err != nil {
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusConflict, errBulkInUse)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"deleted": rows})
}

func (server *Server) bulkDeleteColors(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteColors(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

func (server *Server) bulkDeleteSizes(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteSizes(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

func (server *Server) bulkDeleteAttributeValues(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteAttributeValues(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

func (server *Server) bulkDeleteLoans(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteLoans(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

func (server *Server) bulkDeleteExpenses(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteExpenses(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

func (server *Server) bulkDeleteInventories(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteInventories(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

type bulkDeleteCurrenciesRequest struct {
	Codes []string `json:"codes" binding:"required,min=1,max=500,dive,min=2"`
}

func (server *Server) bulkDeleteCurrencies(ctx *gin.Context) {
	var req bulkDeleteCurrenciesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteCurrencies(ctx, req.Codes)
	server.bulkRespond(ctx, rows, err)
}

type bulkDeleteListItemsURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type bulkDeleteListItemsRequest struct {
	ProductIDs []int64 `json:"product_ids" binding:"required,min=1,max=1000,dive,min=1"`
}

func (server *Server) bulkDeletePriceListItems(ctx *gin.Context) {
	var uri bulkDeleteListItemsURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req bulkDeleteListItemsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeletePriceListItems(ctx, db.DeletePriceListItemsParams{
		PriceListID: uri.ID,
		ProductIds:  req.ProductIDs,
	})
	server.bulkRespond(ctx, rows, err)
}

func (server *Server) bulkDeleteDiscountListItems(ctx *gin.Context) {
	var uri bulkDeleteListItemsURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req bulkDeleteListItemsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteDiscountListItems(ctx, db.DeleteDiscountListItemsParams{
		DiscountListID: uri.ID,
		ProductIds:     req.ProductIDs,
	})
	server.bulkRespond(ctx, rows, err)
}
