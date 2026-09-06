package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
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

// --- Admin-side attendance editing ---------------------------------------
//
// The punch stream is normally append-only and device-authoritative, but an
// admin sometimes has to fix a mislabelled or duplicated row by hand, and
// branches without a biometric device need a way in at all -- hence the
// CSV import below. These endpoints sit under authRoutes (a human operator
// token), never the per-branch key.

type attendanceEventIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateBranchAttendanceEventRequest struct {
	SalespersonName  string `json:"salesperson_name"`
	AttendanceUserID string `json:"attendance_user_id"`
	EventDate        string `json:"event_date" binding:"required"`
	EventTime        string `json:"event_time"`
	EventAt          string `json:"event_at"`
	Type             string `json:"type"`
	Status           string `json:"status"`
}

func (server *Server) updateBranchAttendanceEvent(ctx *gin.Context) {
	var uri attendanceEventIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req updateBranchAttendanceEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.GetBranchAttendanceEvent(ctx, uri.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	event, err := server.store.UpdateBranchAttendanceEvent(ctx, db.UpdateBranchAttendanceEventParams{
		ID:               uri.ID,
		SalespersonName:  req.SalespersonName,
		AttendanceUserID: req.AttendanceUserID,
		EventDate:        req.EventDate,
		EventTime:        req.EventTime,
		EventAt:          req.EventAt,
		Type:             req.Type,
		Status:           req.Status,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"branch_attendance_event": event})
}

func (server *Server) deleteBranchAttendanceEvent(ctx *gin.Context) {
	var uri attendanceEventIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.GetBranchAttendanceEvent(ctx, uri.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := server.store.DeleteBranchAttendanceEvent(ctx, uri.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"message": "attendance event deleted"})
}

// importBranchAttendanceEventsCSV ingests a manual attendance export for a
// branch that has no biometric device wired to the sync pipeline. The
// upload is multipart: a branch_id field plus a `file` whose first row is
// the header `date,time,name,type,status,user_id` (only date is required
// per row). Each row's client_ref is derived from its content
// ("csv-<branchID>-<sha256(row)[:16]>") so re-uploading the same file is
// idempotent, exactly like a replayed device batch.
func (server *Server) importBranchAttendanceEventsCSV(ctx *gin.Context) {
	branchID, err := strconv.ParseInt(ctx.PostForm("branch_id"), 10, 64)
	if err != nil || branchID <= 0 {
		server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("valid branch_id is required"))
		return
	}
	if _, err := server.store.GetBranch(ctx, branchID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, fmt.Errorf("branch not found"))
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("a CSV file is required"))
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("could not parse CSV: %w", err))
		return
	}
	if len(rows) < 2 {
		server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("CSV has no data rows"))
		return
	}

	col := attendanceCSVColumns(rows[0])

	inserted, skipped := 0, 0
	for _, row := range rows[1:] {
		if isBlankRow(row) {
			continue
		}
		date := col.get(row, "date")
		if date == "" {
			server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("row missing a date: %v", row))
			return
		}

		ref := "csv-" + strconv.FormatInt(branchID, 10) + "-" + shortHash(strings.Join(row, "\x1f"))

		if _, err := server.store.GetBranchAttendanceEventByClientRef(ctx, db.GetBranchAttendanceEventByClientRefParams{
			BranchID:  branchID,
			ClientRef: ref,
		}); err == nil {
			skipped++
			continue
		} else if err != sql.ErrNoRows {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}

		eventTime := col.get(row, "time")
		if _, err := server.store.CreateBranchAttendanceEvent(ctx, db.CreateBranchAttendanceEventParams{
			BranchID:         branchID,
			ClientRef:        ref,
			SalespersonName:  col.get(row, "name"),
			AttendanceUserID: col.get(row, "user_id"),
			EventDate:        date,
			EventTime:        eventTime,
			EventAt:          strings.TrimSpace(date + " " + eventTime),
			Type:             col.get(row, "type"),
			Status:           col.get(row, "status"),
		}); err != nil {
			if isUniqueViolation(err) {
				skipped++
				continue
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		inserted++
	}

	if inserted > 0 {
		server.adminHub.broadcastAll(adminWSMessage{
			Type:     "branch_attendance_events",
			BranchID: branchID,
			Amount:   int64(inserted),
		})
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"inserted": inserted, "skipped": skipped})
}

// attendanceCSVColumnIndex maps a canonical column name to its position in
// the uploaded header row, so column order in the file doesn't matter.
type attendanceCSVColumnIndex map[string]int

func attendanceCSVColumns(header []string) attendanceCSVColumnIndex {
	idx := attendanceCSVColumnIndex{}
	for i, name := range header {
		idx[strings.ToLower(strings.TrimSpace(name))] = i
	}
	return idx
}

func (c attendanceCSVColumnIndex) get(row []string, name string) string {
	i, ok := c[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func isBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:16]
}
