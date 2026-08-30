package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// defaultExpenseCategoryIcon / Color fill in when the client omits them, so
// every category row always has something to render.
const (
	defaultExpenseCategoryIcon  = "mdi-tag-outline"
	defaultExpenseCategoryColor = "blue-grey"
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

type createExpenseCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
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
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
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

func (server *Server) listExpenseCategories(ctx *gin.Context) {
	categories, err := server.store.ListExpenseCategories(ctx)
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
	})
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
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
