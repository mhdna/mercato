package api

import (
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/mhdna/kashi/db/sqlc"
)

const (
	// expenseUploadTokenTTL bounds how long a QR code shown in kashi-pos
	// stays scannable -- long enough to walk over and scan it, short
	// enough to keep an unauthenticated upload endpoint's exposure window
	// small.
	expenseUploadTokenTTL       = 30 * time.Minute
	maxExpenseUploadImages      = 3
	maxExpenseUploadImageBytes  = 10 << 20 // 10MB
)

var allowedExpenseImageContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/heic": ".heic",
}

//go:embed expense_upload_assets/upload.html
var expenseUploadPageHTML []byte

// createBranchExpenseUploadToken mints a one-time, expiring token for a
// freshly-created branch expense. Called from createBranchExpense
// (branch_expense.go) only on the genuinely-new-insert path.
func (server *Server) createBranchExpenseUploadToken(ctx *gin.Context, branchExpenseID int64) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate upload token: %w", err)
	}
	token := hex.EncodeToString(raw)

	_, err := server.store.CreateBranchExpenseUploadToken(ctx, db.CreateBranchExpenseUploadTokenParams{
		BranchExpenseID: branchExpenseID,
		Token:           token,
		ExpiresAt:       time.Now().UTC().Add(expenseUploadTokenTTL),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

// expenseUploadPage serves the self-contained mobile upload page. It's a
// static asset (api/expense_upload_assets/upload.html) -- all validity
// checking happens client-side against expenseUploadStatus/
// uploadExpenseImages, so this handler itself never touches the database.
func (server *Server) expenseUploadPage(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", expenseUploadPageHTML)
}

// lookupValidUploadToken resolves a token to its (unexpired) row, treating
// "not found" and "expired" identically to the caller -- a phone scanning a
// stale QR code doesn't need to distinguish the two.
func (server *Server) lookupValidUploadToken(ctx *gin.Context, token string) (db.BranchExpenseUploadToken, error) {
	row, err := server.store.GetBranchExpenseUploadTokenByToken(ctx, token)
	if err != nil {
		return db.BranchExpenseUploadToken{}, err
	}
	if time.Now().UTC().After(row.ExpiresAt) {
		return db.BranchExpenseUploadToken{}, sql.ErrNoRows
	}
	return row, nil
}

func (server *Server) expenseUploadStatus(ctx *gin.Context) {
	token := ctx.Param("token")

	row, err := server.lookupValidUploadToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, fmt.Errorf("this link is no longer valid"))
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	expense, err := server.store.GetBranchExpense(ctx, row.BranchExpenseID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.writeJSON(ctx, http.StatusOK, envelope{
		"description":  expense.Description,
		"amount_label": formatMoneyLabel(expense.Amount, expense.CurrencyCode),
		"used":         row.UsedAt.Valid,
	})
}

// uploadExpenseImages is the only write path a phone with no login can
// reach. It re-validates the token from scratch (unexpired, unused) rather
// than trusting anything the client claimed in an earlier /status call --
// that call is read-only and racy against a second device using the same
// link.
func (server *Server) uploadExpenseImages(ctx *gin.Context) {
	token := ctx.Param("token")

	row, err := server.lookupValidUploadToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, fmt.Errorf("this link is no longer valid"))
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	if row.UsedAt.Valid {
		server.writeError(ctx, http.StatusGone, fmt.Errorf("this link has already been used"))
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	files := form.File["images"]
	if len(files) == 0 {
		server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("no images provided"))
		return
	}
	if len(files) > maxExpenseUploadImages {
		server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("at most %d images allowed", maxExpenseUploadImages))
		return
	}

	expense, err := server.store.GetBranchExpense(ctx, row.BranchExpenseID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	destDir := filepath.Join(server.config.ExpenseImageStorageDir, fmt.Sprintf("%d", expense.ID))
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	for _, fileHeader := range files {
		if fileHeader.Size > maxExpenseUploadImageBytes {
			server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("each image must be under %dMB", maxExpenseUploadImageBytes>>20))
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		ext, ok := allowedExpenseImageContentTypes[contentType]
		if !ok {
			server.writeError(ctx, http.StatusBadRequest, fmt.Errorf("unsupported image type %q", contentType))
			return
		}

		destPath := filepath.Join(destDir, uuid.NewString()+ext)
		if err := saveUploadedExpenseImage(fileHeader, destPath); err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}

		if _, err := server.store.CreateBranchExpenseImage(ctx, db.CreateBranchExpenseImageParams{
			BranchExpenseID: expense.ID,
			FilePath:        destPath,
			ContentType:     contentType,
			SizeBytes:       fileHeader.Size,
		}); err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	if err := server.store.MarkBranchExpenseUploadTokenUsed(ctx, row.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.notify(expense.BranchID, branchWSMessage{
		Type:      "expense_images_uploaded",
		ClientRef: expense.ClientRef,
	})
	server.adminHub.broadcastAll(adminWSMessage{
		Type:     "branch_expense_images_uploaded",
		BranchID: expense.BranchID,
	})

	server.writeJSON(ctx, http.StatusOK, envelope{"success": true})
}

func saveUploadedExpenseImage(fileHeader *multipart.FileHeader, destPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func formatMoneyLabel(amount int64, currencyCode string) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%s%d.%02d %s", sign, amount/100, amount%100, currencyCode)
}
