package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// defaultAssetCategoryIcon / Color fill in when the client omits them, so
// every category row always has something to render. Mirrors the
// loanCategory* / expenseCategory* helpers.
const (
	defaultAssetCategoryIcon  = "mdi-tag-outline"
	defaultAssetCategoryColor = "blue-grey"
)

func assetCategoryIcon(icon string) string {
	if icon == "" {
		return defaultAssetCategoryIcon
	}
	return icon
}

func assetCategoryColor(color string) string {
	if color == "" {
		return defaultAssetCategoryColor
	}
	return color
}

// ---------------------------------------------------------------- assets --

type createAssetRequest struct {
	Code       string    `json:"code" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	CategoryID int64     `json:"category_id" binding:"required,min=1"`
	BoughtAt   time.Time `json:"bought_at"`
}

func (server *Server) createAsset(ctx *gin.Context) {
	var req createAssetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	boughtAt := req.BoughtAt
	if boughtAt.IsZero() {
		boughtAt = time.Now()
	}

	asset, err := server.store.CreateAsset(ctx, db.CreateAssetParams{
		Name:       req.Name,
		Code:       req.Code,
		CategoryID: req.CategoryID,
		BoughtAt:   boughtAt,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, asset)
}

type getAssetRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAsset(ctx *gin.Context) {
	var req getAssetRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	asset, err := server.store.GetAsset(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, asset)
}

type listAssetsRequest struct {
	PageSize   int32  `form:"page_size" binding:"omitempty,min=5,max=100"`
	PageID     int32  `form:"page_id" binding:"min=0"`
	Search     string `form:"search"`
	SortBy     string `form:"sort_by,default=created_at" binding:"oneof=id name code bought_at created_at"`
	SortOrder  string `form:"sort_order,default=desc" binding:"oneof=asc desc"`
	CategoryID int64  `form:"category_id" binding:"omitempty,min=1"`
}

func (server *Server) listAssets(ctx *gin.Context) {
	var req listAssetsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 14
	}

	var categoryID sql.NullInt64
	if req.CategoryID != 0 {
		categoryID = sql.NullInt64{Int64: req.CategoryID, Valid: true}
	}

	assets, err := server.store.ListAssetsPage(ctx, db.ListAssetsPageParams{
		CategoryID: categoryID,
		Search:     req.Search,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		PageSize:   pageSize,
		PageOffset: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountAssets(ctx, db.CountAssetsParams{
		CategoryID: categoryID,
		Search:     req.Search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"assets": assets, "total": total})
}

type updateAssetRequest struct {
	ID         int64     `json:"id" binding:"required,min=1"`
	Code       string    `json:"code" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	CategoryID int64     `json:"category_id" binding:"required,min=1"`
	BoughtAt   time.Time `json:"bought_at"`
}

func (server *Server) updateAsset(ctx *gin.Context) {
	var req updateAssetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	boughtAt := req.BoughtAt
	if boughtAt.IsZero() {
		current, err := server.store.GetAsset(ctx, req.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				server.writeError(ctx, http.StatusNotFound, err)
				return
			}
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		boughtAt = current.BoughtAt
	}

	asset, err := server.store.UpdateAsset(ctx, db.UpdateAssetParams{
		ID:         req.ID,
		Name:       req.Name,
		Code:       req.Code,
		CategoryID: req.CategoryID,
		BoughtAt:   boughtAt,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, asset)
}

type deleteAssetRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAsset(ctx *gin.Context) {
	var req deleteAssetRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteAsset(ctx, req.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (server *Server) bulkDeleteAssets(ctx *gin.Context) {
	var req bulkDeleteIDsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	rows, err := server.store.DeleteAssets(ctx, req.IDs)
	server.bulkRespond(ctx, rows, err)
}

// ------------------------------------------------------- asset categories --

type createAssetCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
}

func (server *Server) createAssetCategory(ctx *gin.Context) {
	var req createAssetCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.CreateAssetCategory(ctx, db.CreateAssetCategoryParams{
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     assetCategoryIcon(req.Icon),
		Color:    assetCategoryColor(req.Color),
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, category)
}

type getAssetCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAssetCategory(ctx *gin.Context) {
	var req getAssetCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.GetAssetCategory(ctx, req.ID)
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

func (server *Server) listAssetCategories(ctx *gin.Context) {
	categories, err := server.store.ListAssetCategories(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, categories)
}

type updateAssetCategoryRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
}

func (server *Server) updateAssetCategory(ctx *gin.Context) {
	var req updateAssetCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	category, err := server.store.UpdateAssetCategory(ctx, db.UpdateAssetCategoryParams{
		ID:       req.ID,
		Name:     req.Name,
		IsActive: req.IsActive,
		Icon:     assetCategoryIcon(req.Icon),
		Color:    assetCategoryColor(req.Color),
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

type deleteAssetCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// deleteAssetCategory hard-deletes a category. assets.category_id has no
// ON DELETE CASCADE, so deleting a category still referenced by an asset
// fails on the FK constraint -- surfaced as a 409, same convention as
// deleteLoanCategory / deleteExpenseCategory.
func (server *Server) deleteAssetCategory(ctx *gin.Context) {
	var req deleteAssetCategoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteAssetCategory(ctx, req.ID); err != nil {
		if isForeignKeyViolation(err) {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
