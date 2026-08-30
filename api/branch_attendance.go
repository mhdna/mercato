package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// attendanceEventBackfillToastCap is the batch size past which a single
// sync run's punches are summarised into one toast rather than one per
// event -- so a first-time backfill of a busy device doesn't storm the
// admin UI.
const attendanceEventBackfillToastCap = 10

// branchAttendanceEventRequest is one row of the batch POSTed to
// /branch/attendance_events (see branchAttendanceEvent in kashi-pos's
// sync_outbox.go). Ref is the device's own "<device_id>|<event_id>" key,
// unique per branch. The time columns are stored as the branch's own text,
// not parsed -- the device is the only authority on their format.
type branchAttendanceEventRequest struct {
	Ref              string `json:"ref" binding:"required"`
	SalespersonName  string `json:"salesperson_name"`
	AttendanceUserID string `json:"attendance_user_id"`
	EventDate        string `json:"event_date" binding:"required"`
	EventTime        string `json:"event_time"`
	EventAt          string `json:"event_at"`
	Type             string `json:"type"`
	Status           string `json:"status"`
}

type branchAttendanceEventsRequest struct {
	Events []branchAttendanceEventRequest `json:"events" binding:"required,min=1,dive"`
}

// createBranchAttendanceEvents ingests one sync run's newly-imported
// punches. Each row is upserted on (branch_id, ref) so replaying the whole
// batch (a retried outbox entry) is idempotent row by row.
func (server *Server) createBranchAttendanceEvents(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchAttendanceEventsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	perEventToasts := len(req.Events) <= attendanceEventBackfillToastCap
	inserted := 0
	for _, event := range req.Events {
		if _, err := server.store.GetBranchAttendanceEventByClientRef(ctx, db.GetBranchAttendanceEventByClientRefParams{
			BranchID:  branchID,
			ClientRef: event.Ref,
		}); err == nil {
			continue
		} else if err != sql.ErrNoRows {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}

		if _, err := server.store.CreateBranchAttendanceEvent(ctx, db.CreateBranchAttendanceEventParams{
			BranchID:         branchID,
			ClientRef:        event.Ref,
			SalespersonName:  event.SalespersonName,
			AttendanceUserID: event.AttendanceUserID,
			EventDate:        event.EventDate,
			EventTime:        event.EventTime,
			EventAt:          event.EventAt,
			Type:             event.Type,
			Status:           event.Status,
		}); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		inserted++

		if perEventToasts {
			server.adminHub.broadcastAll(adminWSMessage{
				Type:     "branch_attendance_event",
				BranchID: branchID,
				Label:    attendanceEventLabel(event),
			})
		}
	}

	if inserted > 0 && !perEventToasts {
		server.adminHub.broadcastAll(adminWSMessage{
			Type:     "branch_attendance_events",
			BranchID: branchID,
			Amount:   int64(inserted),
		})
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"inserted": inserted})
}

func attendanceEventLabel(event branchAttendanceEventRequest) string {
	parts := make([]string, 0, 3)
	if event.SalespersonName != "" {
		parts = append(parts, event.SalespersonName)
	}
	if event.Type != "" {
		parts = append(parts, event.Type)
	}
	if event.EventTime != "" {
		parts = append(parts, event.EventTime)
	}
	return strings.Join(parts, " ")
}

// branchAttendanceChangeRequest is the exact JSON body POSTed to
// /branch/attendance_changes (see branchAttendanceChangePayload in
// kashi-pos's sync_outbox.go). It covers both a cashier-filed change
// request (kind="complaint") and a manager's review of one
// (kind="approval", carrying the approved/rejected Status).
type branchAttendanceChangeRequest struct {
	Ref             string    `json:"ref" binding:"required"`
	Kind            string    `json:"kind" binding:"required,oneof=complaint approval"`
	SalespersonName string    `json:"salesperson_name"`
	AttendanceDate  string    `json:"attendance_date" binding:"required"`
	OriginalTime    string    `json:"original_time"`
	RequestedTime   string    `json:"requested_time"`
	RequestedType   string    `json:"requested_type"`
	Action          string    `json:"action"`
	Status          string    `json:"status"`
	Note            string    `json:"note"`
	Actor           string    `json:"actor"`
	OccurredAt      time.Time `json:"occurred_at" binding:"required"`
}

// createBranchAttendanceChange follows the same idempotent-insert pattern as
// createBranchExpense.
func (server *Server) createBranchAttendanceChange(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchAttendanceChangeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if existing, err := server.store.GetBranchAttendanceChangeByClientRef(ctx, db.GetBranchAttendanceChangeByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.Ref,
	}); err == nil {
		server.writeJSON(ctx, http.StatusOK, envelope{"change": existing})
		return
	} else if err != sql.ErrNoRows {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	change, err := server.store.CreateBranchAttendanceChange(ctx, db.CreateBranchAttendanceChangeParams{
		BranchID:        branchID,
		ClientRef:       req.Ref,
		Kind:            req.Kind,
		SalespersonName: req.SalespersonName,
		AttendanceDate:  req.AttendanceDate,
		OriginalTime:    req.OriginalTime,
		RequestedTime:   req.RequestedTime,
		RequestedType:   req.RequestedType,
		Action:          req.Action,
		Status:          req.Status,
		Note:            req.Note,
		Actor:           req.Actor,
		OccurredAt:      req.OccurredAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			if existing, lookupErr := server.store.GetBranchAttendanceChangeByClientRef(ctx, db.GetBranchAttendanceChangeByClientRefParams{
				BranchID:  branchID,
				ClientRef: req.Ref,
			}); lookupErr == nil {
				server.writeJSON(ctx, http.StatusOK, envelope{"change": existing})
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.adminHub.broadcastAll(adminWSMessage{
		Type:     "branch_attendance_changed",
		BranchID: branchID,
		Kind:     attendanceChangeKind(req),
		Label:    attendanceChangeLabel(req),
	})

	server.writeJSON(ctx, http.StatusOK, envelope{"change": change})
}

// attendanceChangeKind is the discriminator the admin toast keys its
// wording and icon off: "complaint" for a filed request, and the concrete
// "approved"/"rejected" outcome for a manager's review (rather than the
// generic "approval"), since those read very differently.
func attendanceChangeKind(req branchAttendanceChangeRequest) string {
	if req.Kind == "approval" && (req.Status == "approved" || req.Status == "rejected") {
		return req.Status
	}
	return req.Kind
}

func attendanceChangeLabel(req branchAttendanceChangeRequest) string {
	verb := req.Kind
	if req.Kind == "approval" && req.Status != "" {
		verb = req.Status
	}
	parts := make([]string, 0, 3)
	if req.SalespersonName != "" {
		parts = append(parts, req.SalespersonName)
	}
	parts = append(parts, verb)
	if req.AttendanceDate != "" {
		parts = append(parts, req.AttendanceDate)
	}
	return strings.Join(parts, " ")
}

type listBranchAttendanceRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	BranchID int64 `form:"branch_id"`
}

func (server *Server) listBranchAttendanceChanges(ctx *gin.Context) {
	var req listBranchAttendanceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	changes, err := server.store.ListBranchAttendanceChanges(ctx, db.ListBranchAttendanceChangesParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchAttendanceChanges(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_attendance_changes": changes, "total": total})
}

func (server *Server) listBranchAttendanceEvents(ctx *gin.Context) {
	var req listBranchAttendanceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var branchID sql.NullInt64
	if req.BranchID > 0 {
		branchID = sql.NullInt64{Int64: req.BranchID, Valid: true}
	}

	events, err := server.store.ListBranchAttendanceEvents(ctx, db.ListBranchAttendanceEventsParams{
		Limit:    req.PageSize,
		Offset:   req.PageID,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchAttendanceEvents(ctx, branchID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"branch_attendance_events": events, "total": total})
}
