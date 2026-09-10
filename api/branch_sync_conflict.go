package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/token"
	"github.com/sqlc-dev/pqtype"
)

func nullInt64None() sql.NullInt64 { return sql.NullInt64{} }

func rawJSON(b []byte) pqtype.NullRawMessage {
	return pqtype.NullRawMessage{RawMessage: b, Valid: len(b) > 0}
}

// stageBranchSyncConflict upserts a conflict row and pings connected admins.
// Best-effort: a branch push must never fail because staging its conflict
// failed, so errors here are logged, not returned. Returns true when a row
// was written (an already-resolved conflict is left untouched -> false).
func (server *Server) stageBranchSyncConflict(ctx *gin.Context, branchID int64, entity, ref, kind string, payload []byte, centralClientID sql.NullInt64) bool {
	_, err := server.store.UpsertBranchSyncConflict(ctx, db.UpsertBranchSyncConflictParams{
		BranchID:        branchID,
		Entity:          entity,
		Ref:             ref,
		Kind:            kind,
		BranchPayload:   payload,
		CentralClientID: centralClientID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// ON CONFLICT ... WHERE status = 'open' filtered it out: an admin
		// already resolved or dismissed this one. Nothing to do.
		return false
	}
	if err != nil {
		log.Printf("stage branch sync conflict (branch %d, %s %s): %v", branchID, entity, ref, err)
		return false
	}
	server.adminHub.broadcastAll(adminWSMessage{
		Type:     "branch_sync_conflict",
		BranchID: branchID,
		Kind:     kind,
		Label:    entity,
	})
	return true
}

// --- admin-facing API -----------------------------------------------------

type listBranchSyncConflictsRequest struct {
	PageSize  int32  `form:"page_size,default=14" binding:"min=5,max=100"`
	PageID    int32  `form:"page_id,default=0" binding:"min=0"`
	BranchID  int64  `form:"branch_id"`
	Status    string `form:"status"`
	Entity    string `form:"entity"`
	Search    string `form:"search"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

func (server *Server) listBranchSyncConflicts(ctx *gin.Context) {
	var req listBranchSyncConflictsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	rows, err := server.store.ListBranchSyncConflictsPage(ctx, db.ListBranchSyncConflictsPageParams{
		Column1: req.BranchID,
		Column2: req.Status,
		Column3: req.Entity,
		Column4: req.Search,
		Column5: req.SortBy,
		Column6: req.SortOrder,
		Limit:   req.PageSize,
		Offset:  req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	total, err := server.store.CountBranchSyncConflicts(ctx, db.CountBranchSyncConflictsParams{
		Column1: req.BranchID,
		Column2: req.Status,
		Column3: req.Entity,
		Column4: req.Search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"conflicts": rows, "total": total})
}

type resolveBranchSyncConflictRequest struct {
	// Choice meanings by conflict kind:
	//   phone_name_mismatch: "central" (keep kashi's name) | "branch" (take
	//     the till's) | "custom" (use Name)
	//   invalid_phone: "merge" (link to ClientID) | "create" (new client
	//     with Phone)
	//   unknown_product: "create" (adopt the reported product into the
	//     catalog)
	Choice   string `json:"choice" binding:"required"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	ClientID int64  `json:"client_id"`
}

func (server *Server) resolveBranchSyncConflict(ctx *gin.Context) {
	var idReq branchInvoiceIDRequest // {ID int64 uri:"id" binding:"required,min=1"}
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req resolveBranchSyncConflictRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	conflict, err := server.store.GetBranchSyncConflict(ctx, idReq.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	if conflict.Status != "open" {
		server.writeError(ctx, http.StatusConflict, errors.New("conflict is already resolved"))
		return
	}

	if err := server.applyConflictResolution(ctx, conflict, req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	resolvedBy := server.currentUserID(ctx)
	resolutionJSON, _ := json.Marshal(req)
	updated, err := server.store.ResolveBranchSyncConflict(ctx, db.ResolveBranchSyncConflictParams{
		ID:         conflict.ID,
		Status:     "resolved",
		Resolution: rawJSON(resolutionJSON),
		ResolvedBy: resolvedBy,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Let the branch pull the agreed record back down.
	server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
	server.writeJSON(ctx, http.StatusOK, envelope{"conflict": updated})
}

func (server *Server) dismissBranchSyncConflict(ctx *gin.Context) {
	var idReq branchInvoiceIDRequest
	if err := ctx.ShouldBindUri(&idReq); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	updated, err := server.store.ResolveBranchSyncConflict(ctx, db.ResolveBranchSyncConflictParams{
		ID:         idReq.ID,
		Status:     "dismissed",
		Resolution: rawJSON([]byte(`{"choice":"dismissed"}`)),
		ResolvedBy: server.currentUserID(ctx),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"conflict": updated})
}

type bulkResolveRequest struct {
	IDs    []int64 `json:"ids" binding:"required,min=1"`
	Choice string  `json:"choice" binding:"required"`
}

// bulkResolveBranchSyncConflicts applies the same choice to many conflicts
// -- the "take the branch's spelling for all of these" case. Only choices
// that need no per-row parameter are accepted here (central / branch /
// dismissed); anything needing a Name/Phone/ClientID must go one at a time.
func (server *Server) bulkResolveBranchSyncConflicts(ctx *gin.Context) {
	var req bulkResolveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	switch req.Choice {
	case "central", "branch", "dismissed":
	default:
		server.writeError(ctx, http.StatusBadRequest, errors.New("bulk resolve accepts only central, branch or dismissed"))
		return
	}

	resolvedBy := server.currentUserID(ctx)
	resolved := 0
	for _, id := range req.IDs {
		conflict, err := server.store.GetBranchSyncConflict(ctx, id)
		if err != nil || conflict.Status != "open" {
			continue
		}
		status := "resolved"
		if req.Choice == "dismissed" {
			status = "dismissed"
		} else if err := server.applyConflictResolution(ctx, conflict, resolveBranchSyncConflictRequest{Choice: req.Choice}); err != nil {
			continue
		}
		resJSON, _ := json.Marshal(map[string]string{"choice": req.Choice})
		if _, err := server.store.ResolveBranchSyncConflict(ctx, db.ResolveBranchSyncConflictParams{
			ID:         id,
			Status:     status,
			Resolution: rawJSON(resJSON),
			ResolvedBy: resolvedBy,
		}); err == nil {
			resolved++
		}
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "client_updated"})
	server.writeJSON(ctx, http.StatusOK, envelope{"resolved": resolved})
}

// applyConflictResolution performs the side effect a resolution choice
// implies (renaming a client, writing a link, adopting a product). It does
// NOT flip the conflict's status -- the caller does that once this succeeds.
func (server *Server) applyConflictResolution(ctx *gin.Context, conflict db.BranchSyncConflict, req resolveBranchSyncConflictRequest) error {
	switch conflict.Kind {
	case "phone_name_mismatch":
		if req.Choice == "central" {
			return nil // keep kashi's record untouched
		}
		if !conflict.CentralClientID.Valid {
			return errors.New("conflict has no central client to update")
		}
		var bp branchClientRequest
		if err := json.Unmarshal(conflict.BranchPayload, &bp); err != nil {
			return err
		}
		name := bp.Name
		if req.Choice == "custom" {
			if strings.TrimSpace(req.Name) == "" {
				return errors.New("a custom name is required")
			}
			name = strings.TrimSpace(req.Name)
		} else if req.Choice != "branch" {
			return errors.New("choice must be central, branch or custom")
		}
		current, err := server.store.GetClient(ctx, conflict.CentralClientID.Int64)
		if err != nil {
			return err
		}
		_, err = server.store.UpdateClient(ctx, db.UpdateClientParams{
			ID:         current.ID,
			Name:       name,
			Phone:      current.Phone,
			ClientType: current.ClientType,
		})
		return err

	case "invalid_phone":
		var bp branchClientRequest
		if err := json.Unmarshal(conflict.BranchPayload, &bp); err != nil {
			return err
		}
		branchClientID, err := strconv.ParseInt(conflict.Ref, 10, 64)
		if err != nil {
			return err
		}
		var targetID int64
		switch req.Choice {
		case "merge":
			if req.ClientID <= 0 {
				return errors.New("client_id is required to merge")
			}
			if _, err := server.store.GetClient(ctx, req.ClientID); err != nil {
				return errors.New("target client not found")
			}
			targetID = req.ClientID
		case "create":
			if !validBranchPhone(req.Phone) {
				return errors.New("a valid phone is required to create the client")
			}
			created, err := server.store.CreateClient(ctx, db.CreateClientParams{
				Name:       bp.Name,
				Phone:      strings.TrimSpace(req.Phone),
				ClientType: "retail",
			})
			if err != nil {
				return err
			}
			targetID = created.ID
		default:
			return errors.New("choice must be merge or create")
		}
		return server.store.UpsertClientLink(ctx, db.UpsertClientLinkParams{
			BranchID:       conflict.BranchID,
			BranchClientID: branchClientID,
			ClientID:       targetID,
		})

	case "unknown_product":
		if req.Choice != "create" {
			return errors.New("choice must be create")
		}
		return server.adoptBranchProduct(ctx, conflict)

	default:
		return errors.New("unknown conflict kind")
	}
}

// adoptBranchProduct creates the products + product_variants rows for a
// barcode a branch reported that kashi didn't have. Colour/size are
// resolved by name, created if absent; brand/kind/season/year from the
// branch payload are left for the admin to fill in kashi's product editor
// (they live in the separate attribute system).
func (server *Server) adoptBranchProduct(ctx *gin.Context, conflict db.BranchSyncConflict) error {
	var p branchProductReportItem
	if err := json.Unmarshal(conflict.BranchPayload, &p); err != nil {
		return err
	}
	if p.Barcode == "" {
		return errors.New("reported product has no barcode")
	}
	if _, err := server.store.GetProductVariantByBarcode(ctx, p.Barcode); err == nil {
		return nil // adopted already (double resolve)
	}

	code := strings.TrimSpace(p.Code)
	if code == "" {
		code = p.Barcode
	}
	product, err := server.store.GetProductByCode(ctx, code)
	if errors.Is(err, sql.ErrNoRows) {
		product, err = server.store.CreateProduct(ctx, db.CreateProductParams{
			Code:        code,
			Name:        firstNonEmpty(p.Name, code),
			Description: p.Description,
		})
	}
	if err != nil {
		return err
	}

	colorID, err := server.resolveColorID(ctx, p.Color)
	if err != nil {
		return err
	}
	sizeID, err := server.resolveSizeID(ctx, p.Size)
	if err != nil {
		return err
	}
	price := sql.NullInt64{}
	if p.Price != 0 {
		price = sql.NullInt64{Int64: p.Price, Valid: true}
	}
	_, err = server.store.CreateProductVariant(ctx, db.CreateProductVariantParams{
		ProductID: product.ID,
		ColorID:   colorID,
		SizeID:    sizeID,
		Barcode:   p.Barcode,
		Price:     price,
	})
	return err
}

func (server *Server) resolveColorID(ctx *gin.Context, name string) (sql.NullInt64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return sql.NullInt64{}, nil
	}
	if c, err := server.store.GetColorByName(ctx, name); err == nil {
		return sql.NullInt64{Int64: c.ID, Valid: true}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, err
	}
	c, err := server.store.CreateColor(ctx, db.CreateColorParams{Name: name, HexValue: ""})
	if err != nil {
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{Int64: c.ID, Valid: true}, nil
}

func (server *Server) resolveSizeID(ctx *gin.Context, name string) (sql.NullInt64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return sql.NullInt64{}, nil
	}
	if s, err := server.store.GetSizeByName(ctx, name); err == nil {
		return sql.NullInt64{Int64: s.ID, Valid: true}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, err
	}
	s, err := server.store.CreateSize(ctx, db.CreateSizeParams{Name: name, Type: "other", Order: "0"})
	if err != nil {
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{Int64: s.ID, Valid: true}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// currentUserID resolves the PASETO username to a users.id for the audit
// columns, or a NULL when it can't (never worth failing a resolve over).
func (server *Server) currentUserID(ctx *gin.Context) sql.NullInt64 {
	payload, ok := ctx.MustGet(authoizationPayloadKey).(*token.Payload)
	if !ok {
		return sql.NullInt64{}
	}
	user, err := server.store.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: user.ID, Valid: true}
}
