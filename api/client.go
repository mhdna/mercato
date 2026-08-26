package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createClientRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	// Optional; defaults to "retail" like the column itself.
	ClientType string `json:"client_type" binding:"omitempty,oneof=retail wholesale"`
}

func (server *Server) createClient(ctx *gin.Context) {
	var req createClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if req.ClientType == "" {
		req.ClientType = "retail"
	}

	arg := db.CreateClientParams{
		Name:       req.Name,
		Phone:      req.Phone,
		ClientType: req.ClientType,
	}

	client, err := server.store.CreateClient(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
	ctx.JSON(http.StatusOK, client)
}

type getClientRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getClient(ctx *gin.Context) {
	var req getClientRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	client, err := server.store.GetClient(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, client)
}

type listClientsRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	// Optional; filters to one tab's worth of clients on the admin page.
	// Omitted (empty) means all types.
	ClientType string `form:"client_type" binding:"omitempty,oneof=retail wholesale"`
}

func (server *Server) listClients(ctx *gin.Context) {
	var req listClientsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	clientType := sql.NullString{String: req.ClientType, Valid: req.ClientType != ""}
	arg := db.ListClientsParams{
		Limit:      req.PageSize,
		Offset:     req.PageID,
		ClientType: clientType,
	}
	clients, err := server.store.ListClients(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountClients(ctx, clientType)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"clients": clients, "total": total})
}

type listClientsBySpendingRequest struct {
	PageSize   int32  `form:"page_size,default=20" binding:"min=1,max=100"`
	PageID     int32  `form:"page_id,default=0" binding:"min=0"`
	ClientType string `form:"client_type" binding:"omitempty,oneof=retail wholesale"`
	BranchID   int64  `form:"branch_id"`
	Search     string `form:"search"`
	// What to rank clients by; see the ORDER BY in ListClientsBySpending.
	SortBy string `form:"sort_by,default=spending" binding:"omitempty,oneof=spending invoices items"`
}

// listClientsBySpending is the clients loyalty-pyramid page's data source:
// every client ranked by total_spent, invoice_count, or item_count (see
// ListClientsBySpending) rather than the alphabetical order the plain
// clients list/table uses.
func (server *Server) listClientsBySpending(ctx *gin.Context) {
	var req listClientsBySpendingRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	clientType := sql.NullString{String: req.ClientType, Valid: req.ClientType != ""}
	search := sql.NullString{String: req.Search, Valid: req.Search != ""}
	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	clients, err := server.store.ListClientsBySpending(ctx, db.ListClientsBySpendingParams{
		Limit:      req.PageSize,
		Offset:     req.PageID,
		BranchID:   branchID,
		ClientType: clientType,
		Search:     search,
		SortBy:     req.SortBy,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountClientsBySpending(ctx, db.CountClientsBySpendingParams{
		ClientType: clientType,
		Search:     search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"clients": clients, "total": total})
}

type updateClientRequest struct {
	ID    int64  `json:"id" binding:"required,min=1"`
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	// Optional; blank keeps the client's current type rather than resetting
	// it to "retail".
	ClientType string `json:"client_type" binding:"omitempty,oneof=retail wholesale"`
}

func (server *Server) updateClient(ctx *gin.Context) {
	var req updateClientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	clientType := req.ClientType
	if clientType == "" {
		existing, err := server.store.GetClient(ctx, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				server.writeError(ctx, http.StatusBadRequest, err)
				return
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		clientType = existing.ClientType
	}

	arg := db.UpdateClientParams{
		ID:         req.ID,
		Name:       req.Name,
		Phone:      req.Phone,
		ClientType: clientType,
	}
	client, err := server.store.UpdateClient(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
	ctx.JSON(http.StatusOK, client)
}

type deleteClientRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteClient(ctx *gin.Context) {
	var req deleteClientRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteClient(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

type clientLoyaltyTotalRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getClientLoyaltyTotal is the redemption-time read: how many loyalty
// points this client has earned across every branch they've ever bought
// from, not just kashi's own directly-created invoices. See
// SumClientLoyaltyPoints in db/query/client.sql.
func (server *Server) getClientLoyaltyTotal(ctx *gin.Context) {
	var req clientLoyaltyTotalRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	total, err := server.store.SumClientLoyaltyPoints(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"client_id": req.ID, "loyalty_points_total": total})
}
