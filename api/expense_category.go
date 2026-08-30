package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// defaultExpenseCategoryIcon / Color / Scope fill in when the client omits
// them, so every category row always has something to render and a valid
// scope. 'branch'-scoped categories are the ones synced down to kashi-pos.
const (
	defaultExpenseCategoryIcon  = "mdi-tag-outline"
	defaultExpenseCategoryColor = "blue-grey"
	defaultExpenseCategoryScope = "central"
)

func expenseCategoryIcon(icon string) string {
	if icon == "" {
		return defaultExpenseCategoryIcon
	}
	return icon
}

func expenseCategoryColor(color string) string {
	if color == "" {
		return defaultExpenseCategoryColor
	}
	return color
}

func expenseCategoryScope(scope string) string {
	if scope != "branch" {
		return defaultExpenseCategoryScope
	}
	return scope
}

type createExpenseCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Scope    string `json:"scope"`
}

func (server *Server) createExpenseCategory(ctx *gin.Context) {
	var req createExpenseCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.CreateExpenseCategory(ctx, db.CreateExpenseCategoryParams{
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     expenseCategoryIcon(req.Icon),
		Color:    expenseCategoryColor(req.Color),
		Scope:    expenseCategoryScope(req.Scope),
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "expense_categories_updated"})
	ctx.JSON(http.StatusOK, category)
}

type getExpenseCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getExpenseCategory(ctx *gin.Context) {
	var req getExpenseCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.GetExpenseCategory(ctx, req.ID)
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

type listExpenseCategoriesRequest struct {
	Scope string `form:"scope" binding:"omitempty,oneof=central branch"`
}

func (server *Server) listExpenseCategories(ctx *gin.Context) {
	var req listExpenseCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var scope sql.NullString
	if req.Scope != "" {
		scope = sql.NullString{String: req.Scope, Valid: true}
	}

	categories, err := server.store.ListExpenseCategories(ctx, scope)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, categories)
}

type updateExpenseCategoryRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Scope    string `json:"scope"`
}

func (server *Server) updateExpenseCategory(ctx *gin.Context) {
	var req updateExpenseCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.UpdateExpenseCategory(ctx, db.UpdateExpenseCategoryParams{
		ID:       req.ID,
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     expenseCategoryIcon(req.Icon),
		Color:    expenseCategoryColor(req.Color),
		Scope:    expenseCategoryScope(req.Scope),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "expense_categories_updated"})
	ctx.JSON(http.StatusOK, category)
}

type deleteExpenseCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// deleteExpenseCategory hard-deletes a category. expenses/branch_expenses/
// recurring_expenses.category_id has no ON DELETE CASCADE, so deleting a
// category still referenced by any of them fails on the FK constraint --
// surfaced as a 409, same convention as deleteLoanCategory.
func (server *Server) deleteExpenseCategory(ctx *gin.Context) {
	var req deleteExpenseCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteExpenseCategory(ctx, req.ID); err != nil {
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "expense_categories_updated"})
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
