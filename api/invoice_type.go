package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createInvoiceTypeRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	IsDefault bool   `json:"is_default"`
	IsActive  bool   `json:"is_active"`
}

func (server *Server) createInvoiceType(ctx *gin.Context) {
	var req createInvoiceTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateInvoiceTypeParams{
		Name:      req.Name,
		Code:      req.Code,
		IsDefault: req.IsDefault,
		IsActive:  req.IsActive,
	}

	invoiceType, err := server.store.CreateInvoiceType(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, invoiceType)
}

type getInvoiceTypeRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getInvoiceType(ctx *gin.Context) {
	var req getInvoiceTypeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	invoiceType, err := server.store.GetInvoiceType(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, invoiceType)
}

func (server *Server) listInvoiceTypes(ctx *gin.Context) {
	invoiceTypes, err := server.store.ListInvoiceTypes(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, invoiceTypes)
}

type updateInvoiceTypeRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (server *Server) updateInvoiceType(ctx *gin.Context) {
	var req updateInvoiceTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.UpdateInvoiceTypeParams{
		ID:       req.ID,
		Name:     req.Name,
		Code:     req.Code,
		IsActive: req.IsActive,
	}

	invoiceType, err := server.store.UpdateInvoiceType(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, invoiceType)
}
