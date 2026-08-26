package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type listBranchExpenseImagesRequest struct {
	PageSize int32 `form:"page_size,default=20" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

// listBranchExpenseImages feeds the admin Storage explorer -- every
// uploaded receipt photo, newest first, with just enough of its parent
// expense/branch inlined that the UI never needs a second round trip per
// row (see ListBranchExpenseImages's join in db/query/branch_expense_image.sql).
func (server *Server) listBranchExpenseImages(ctx *gin.Context) {
	var req listBranchExpenseImagesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	images, err := server.store.ListBranchExpenseImages(ctx, db.ListBranchExpenseImagesParams{
		Limit:  req.PageSize,
		Offset: req.PageID * req.PageSize,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountBranchExpenseImages(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"images": images, "total": total})
}

type listBranchExpenseImagesForExpenseRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) listBranchExpenseImagesForExpense(ctx *gin.Context) {
	var req listBranchExpenseImagesForExpenseRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	images, err := server.store.ListBranchExpenseImagesForExpense(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"images": images})
}

type getBranchExpenseImageFileRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getBranchExpenseImageFile streams a receipt photo from disk. Kept under
// authRoutes (Bearer-token gated) like the rest of the admin API, rather
// than a plain static file route -- an <img> tag can't send an
// Authorization header, so the admin UI fetches this as a blob and renders
// it via an object URL instead of a bare <img src>.
func (server *Server) getBranchExpenseImageFile(ctx *gin.Context) {
	var req getBranchExpenseImageFileRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	image, err := server.store.GetBranchExpenseImage(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusNotFound, err)
		return
	}

	ctx.Header("Content-Type", image.ContentType)
	ctx.File(image.FilePath)
}

type branchExpenseUploadStatusRequest struct {
	ClientRef string `form:"client_ref" binding:"required"`
}

// branchExpenseUploadStatus is kashi-pos's poll-based fallback while its QR
// dialog is open, in case the "expense_images_uploaded" WS push (see
// uploadExpenseImages in expense_upload.go) never arrives -- the branch may
// simply not have a live connection at that moment. Mirrors the
// poll-alongside-push pattern branchHub.notify's doc comment describes for
// every other pushed entity.
func (server *Server) branchExpenseUploadStatus(ctx *gin.Context) {
	branchID := ctx.MustGet(branchIDKey).(int64)

	var req branchExpenseUploadStatusRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	expense, err := server.store.GetBranchExpenseByClientRef(ctx, db.GetBranchExpenseByClientRefParams{
		BranchID:  branchID,
		ClientRef: req.ClientRef,
	})
	if err != nil {
		server.writeError(ctx, http.StatusNotFound, err)
		return
	}

	images, err := server.store.ListBranchExpenseImagesForExpense(ctx, expense.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{"images_uploaded": len(images) > 0})
}
