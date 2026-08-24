package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

type Attributes struct {
	Category    string `json:"category"`
	SubCategory string `json:"subcategory"`
	Brand       string `json:"brand"`
	Kind        string `json:"kind"`
	Type        string `json:"type"`
	Unit        string `json:"unit"`
	Year        string `json:"year"`
	Season      string `json:"season"`
	Origin      string `json:"origin"`
}

type createProductRequest struct {
	Code        string     `json:"code" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Price       int64      `json:"price"`
	Discount    int16      `json:"discount"`
	Attributes  Attributes `json:"attributes"`
	Barcode     *int64     `json:"barcode"`
	ColorID     *int64     `json:"color_id"`
	SizeID      *int64     `json:"size_id"`
}

func (server *Server) createProduct(ctx *gin.Context) {
	var req createProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	// TODO: refactor to be more clear (for now we have to enter each ID manually)
	attributeValues := []db.AttributesValue{
		{AttributeID: 0, Value: req.Attributes.Category},
		{AttributeID: 1, Value: req.Attributes.SubCategory},
		{AttributeID: 2, Value: req.Attributes.Brand},
		{AttributeID: 3, Value: req.Attributes.Kind},
		{AttributeID: 4, Value: req.Attributes.Type},
		{AttributeID: 5, Value: req.Attributes.Unit},
		{AttributeID: 6, Value: req.Attributes.Year},
		{AttributeID: 7, Value: req.Attributes.Season},
		{AttributeID: 8, Value: req.Attributes.Origin},
	}

	createProductArg := db.CreateProductTxParams{
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		Discount:        req.Discount,
		AttributeValues: attributeValues,
		Barcode:         req.Barcode,
		ColorID:         req.ColorID,
		SizeID:          req.SizeID,
	}

	product, err := server.store.CreateProductTx(ctx, createProductArg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				server.writeError(ctx, http.StatusForbidden, err)
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	ctx.JSON(http.StatusOK, product)
}

type updateProductRequest struct {
	ID          int64  `json:"id" binding:"required,min=1"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// updateProduct edits the shared product-level fields (name/code/
// description) — not any variant's own price/barcode/color/size, which is
// updateProductVariant's job below. The two are separate endpoints because
// they're separate rows now: a product can have several variants, and
// editing "the product" shouldn't require picking one of them.
func (server *Server) updateProduct(ctx *gin.Context) {
	var req updateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.UpdateProduct(ctx, db.UpdateProductParams{
		ID:          req.ID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	product, err := server.store.GetProduct(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	ctx.JSON(http.StatusOK, product)
}

type updateProductVariantRequest struct {
	ID       int64  `json:"id" binding:"required,min=1"`
	ColorID  *int64 `json:"color_id"`
	SizeID   *int64 `json:"size_id"`
	Barcode  string `json:"barcode" binding:"required"`
	Price    *int64 `json:"price"`
	IsActive bool   `json:"is_active"`
}

func (server *Server) updateProductVariant(ctx *gin.Context) {
	var req updateProductVariantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	colorID := sql.NullInt64{}
	if req.ColorID != nil {
		colorID = sql.NullInt64{Int64: *req.ColorID, Valid: true}
	}
	sizeID := sql.NullInt64{}
	if req.SizeID != nil {
		sizeID = sql.NullInt64{Int64: *req.SizeID, Valid: true}
	}
	price := sql.NullInt64{}
	if req.Price != nil {
		price = sql.NullInt64{Int64: *req.Price, Valid: true}
	}

	variant, err := server.store.UpdateProductVariant(ctx, db.UpdateProductVariantParams{
		ID:       req.ID,
		ColorID:  colorID,
		SizeID:   sizeID,
		Barcode:  req.Barcode,
		Price:    price,
		IsActive: req.IsActive,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			server.writeError(ctx, http.StatusForbidden, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	server.writeJSON(ctx, http.StatusOK, envelope{"variant": variant})
}

type getProductRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getProduct(ctx *gin.Context) {
	var req getProductRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	product, err := server.store.GetProduct(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	attributes, err := server.store.GetProductAttributes(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// TODO: use envelop
	ctx.JSON(http.StatusOK, gin.H{
		"product":    product,
		"attributes": attributes,
	})
}

type listProductRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listProducts(ctx *gin.Context) {
	var req listProductRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListProductsParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	products, err := server.store.ListProducts(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountProducts(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    total,
	})
}

type deleteProductRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteProduct(ctx *gin.Context) {
	var req deleteProductRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteProduct(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
