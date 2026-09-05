package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// Coupon categories are an admin-only taxonomy referenced by
// coupons.category_id. They mirror loan_categories (see loan_category.go)
// but are central-only -- there is no branch sync -- so no branchHub
// broadcast on write. defaultCouponCategory* fill in when the client omits
// a field so every row always has an icon, colour and a valid scope.
const (
	defaultCouponCategoryIcon  = "mdi-tag-outline"
	defaultCouponCategoryColor = "blue-grey"
	defaultCouponCategoryScope = "central"
)

func couponCategoryIcon(icon string) string {
	if icon == "" {
		return defaultCouponCategoryIcon
	}
	return icon
}

func couponCategoryColor(color string) string {
	if color == "" {
		return defaultCouponCategoryColor
	}
	return color
}

func couponCategoryScope(scope string) string {
	if scope != "branch" {
		return defaultCouponCategoryScope
	}
	return scope
}

type createCouponCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Scope    string `json:"scope"`
}

func (server *Server) createCouponCategory(ctx *gin.Context) {
	var req createCouponCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.CreateCouponCategory(ctx, db.CreateCouponCategoryParams{
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     couponCategoryIcon(req.Icon),
		Color:    couponCategoryColor(req.Color),
		Scope:    couponCategoryScope(req.Scope),
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, category)
}

type getCouponCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getCouponCategory(ctx *gin.Context) {
	var req getCouponCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.GetCouponCategory(ctx, req.ID)
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

type listCouponCategoriesRequest struct {
	Scope string `form:"scope" binding:"omitempty,oneof=central branch"`
}

func (server *Server) listCouponCategories(ctx *gin.Context) {
	var req listCouponCategoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var scope sql.NullString
	if req.Scope != "" {
		scope = sql.NullString{String: req.Scope, Valid: true}
	}

	categories, err := server.store.ListCouponCategories(ctx, scope)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, categories)
}

type updateCouponCategoryRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Scope    string `json:"scope"`
}

func (server *Server) updateCouponCategory(ctx *gin.Context) {
	var req updateCouponCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.UpdateCouponCategory(ctx, db.UpdateCouponCategoryParams{
		ID:       req.ID,
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     couponCategoryIcon(req.Icon),
		Color:    couponCategoryColor(req.Color),
		Scope:    couponCategoryScope(req.Scope),
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

type deleteCouponCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// deleteCouponCategory hard-deletes a category. coupons.category_id has no
// ON DELETE CASCADE, so deleting one still referenced by a coupon fails on
// the FK constraint -- surfaced as a 409 so the admin UI can suggest
// deactivating the category instead.
func (server *Server) deleteCouponCategory(ctx *gin.Context) {
	var req deleteCouponCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteCouponCategory(ctx, req.ID); err != nil {
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, envelope{"deleted": true})
}
