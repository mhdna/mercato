package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// listAuditLogs is always server-paginated -- unlike most list endpoints,
// this table grows without bound (every audited write logs a row).
type listAuditLogsRequest struct {
	EntityType string `form:"entity_type"`
	EntityID   int64  `form:"entity_id"`
	Action     string `form:"action"`
	ActorID    int64  `form:"actor_id"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	PageSize   int32  `form:"page_size,default=50" binding:"min=1,max=500"`
	PageID     int32  `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listAuditLogs(ctx *gin.Context) {
	var req listAuditLogsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var entityType sql.NullString
	if req.EntityType != "" {
		entityType = sql.NullString{String: req.EntityType, Valid: true}
	}
	var entityID, actorID sql.NullInt64
	if req.EntityID > 0 {
		entityID = sql.NullInt64{Int64: req.EntityID, Valid: true}
	}
	if req.ActorID > 0 {
		actorID = sql.NullInt64{Int64: req.ActorID, Valid: true}
	}
	var action db.NullAuditAction
	if req.Action != "" {
		action = db.NullAuditAction{AuditAction: db.AuditAction(req.Action), Valid: true}
	}
	fromDate, err := parseOptionalDate(req.FromDate)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	toDate, err := parseOptionalDate(req.ToDate)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	respondList(server, ctx, "audit_logs",
		func() ([]db.ListAuditLogsRow, error) {
			return server.store.ListAuditLogs(ctx, db.ListAuditLogsParams{
				EntityType: entityType,
				EntityID:   entityID,
				Action:     action,
				ActorID:    actorID,
				FromDate:   fromDate,
				ToDate:     toDate,
				PageLimit:  req.PageSize,
				PageOffset: req.PageID,
			})
		},
		func() (int64, error) {
			return server.store.CountAuditLogs(ctx, db.CountAuditLogsParams{
				EntityType: entityType,
				EntityID:   entityID,
				Action:     action,
				ActorID:    actorID,
				FromDate:   fromDate,
				ToDate:     toDate,
			})
		},
	)
}
