package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// defaultLoanCategoryIcon / Color / Scope fill in when the client omits
// them, so every category row always has something to render and a valid
// scope. Mirrors the expenseCategory* helpers in expense_category.go.
const (
	defaultLoanCategoryIcon  = "mdi-tag-outline"
	defaultLoanCategoryColor = "blue-grey"
	defaultLoanCategoryScope = "central"
)

func loanCategoryIcon(icon string) string {
	if icon == "" {
		return defaultLoanCategoryIcon
	}
	return icon
}

func loanCategoryColor(color string) string {
	if color == "" {
		return defaultLoanCategoryColor
	}
	return color
}

func loanCategoryScope(scope string) string {
	if scope != "branch" {
		return defaultLoanCategoryScope
	}
	return scope
}

type createLoanCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Scope    string `json:"scope"`
}

func (server *Server) createLoanCategory(ctx *gin.Context) {
	var req createLoanCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.CreateLoanCategory(ctx, db.CreateLoanCategoryParams{
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     loanCategoryIcon(req.Icon),
		Color:    loanCategoryColor(req.Color),
		Scope:    loanCategoryScope(req.Scope),
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "loan_categories_updated"})
	ctx.JSON(http.StatusOK, category)
}

type getLoanCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getLoanCategory(ctx *gin.Context) {
	var req getLoanCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.GetLoanCategory(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, category)
}

type listLoanCategoriesRequest struct {
	Scope string `form:"scope" binding:"omitempty,oneof=central branch"`
}

func (server *Server) listLoanCategories(ctx *gin.Context) {
	var req listLoanCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var scope sql.NullString
	if req.Scope != "" {
		scope = sql.NullString{String: req.Scope, Valid: true}
	}

	categories, err := server.store.ListLoanCategories(ctx, scope)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, categories)
}

type updateLoanCategoryRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Scope    string `json:"scope"`
}

func (server *Server) updateLoanCategory(ctx *gin.Context) {
	var req updateLoanCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.UpdateLoanCategory(ctx, db.UpdateLoanCategoryParams{
		ID:       req.ID,
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     loanCategoryIcon(req.Icon),
		Color:    loanCategoryColor(req.Color),
		Scope:    loanCategoryScope(req.Scope),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "loan_categories_updated"})
	ctx.JSON(http.StatusOK, category)
}

type deleteLoanCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// deleteLoanCategory hard-deletes a category. loans.category_id has no
// ON DELETE CASCADE, so deleting a category still referenced by a loan
// fails on the FK constraint -- surfaced as a 409 rather than a generic
// 500 so the admin UI can explain "deactivate instead of delete" here.
func (server *Server) deleteLoanCategory(ctx *gin.Context) {
	var req deleteLoanCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteLoanCategory(ctx, req.ID); err != nil {
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "loan_categories_updated"})
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
