package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createSupplierRequest struct {
	Name             string  `json:"name" binding:"required"`
	Phone            string  `json:"phone" binding:"required"`
	Country          string  `json:"country" binding:"required"`
	Address          string  `json:"address" binding:"required"`
	AddressLatitude  float64 `json:"latitude" binding:"required"`
	AddressLongitude float64 `json:"longitude" binding:"required"`
}

func (server *Server) createSupplier(ctx *gin.Context) {
	var req createSupplierRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateSupplierParams{
		Name:    req.Name,
		Phone:   req.Phone,
		Country: req.Country,
		Address: req.Address,
		// TODO: refactor this
		AddressLatitude:  sql.NullFloat64{Float64: req.AddressLongitude, Valid: true},
		AddressLongitude: sql.NullFloat64{Float64: req.AddressLatitude, Valid: true},
	}

	supplier, err := server.store.CreateSupplier(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, supplier)
}

type getSupplierRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getSupplier(ctx *gin.Context) {
	var req getSupplierRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	supplier, err := server.store.GetSupplier(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, supplier)
}

type listSupplierRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listSuppliers(ctx *gin.Context) {
	var req listSupplierRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListSuppliersParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	suppliers, err := server.store.ListSuppliers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	total, err := server.store.CountSuppliers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"suppliers": suppliers,
		"total":     total,
	})
}

// type updateSupplierRequest struct {
// 	ID    int64  `json:"id" binding:"required,min=1"`
// 	Name  string `json:"name" binding:"required"`
// 	Phone string `json:"Phone" binding:"required"`
// }

// func (server *Server) updateSupplier(ctx *gin.Context) {
// 	var req updateClientRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	arg := db.UpdateClientParams{
// 		ID:    req.ID,
// 		Name:  req.Name,
// 		Phone: req.Name,
// 	}
// 	client, err := server.store.UpdateClient(ctx, arg)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}
// 	ctx.JSON(http.StatusOK, client)
// }
