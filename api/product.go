package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

// attributeValues maps the fixed 9 attribute names to their seeded ids
// (0-8, per 000005_create_products_attributes_table.up.sql). Shared by
// create and update so the mapping lives in one place.
func attributeValues(a Attributes) []db.AttributesValue {
	return []db.AttributesValue{
		{AttributeID: 0, Value: a.Category},
		{AttributeID: 1, Value: a.SubCategory},
		{AttributeID: 2, Value: a.Brand},
		{AttributeID: 3, Value: a.Kind},
		{AttributeID: 4, Value: a.Type},
		{AttributeID: 5, Value: a.Unit},
		{AttributeID: 6, Value: a.Year},
		{AttributeID: 7, Value: a.Season},
		{AttributeID: 8, Value: a.Origin},
	}
}

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
	// When creating several variants at once: one variant per colour x size.
	ColorIDs []int64 `json:"color_ids"`
	SizeIDs  []int64 `json:"size_ids"`
}

func (server *Server) createProduct(ctx *gin.Context) {
	var req createProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	createProductArg := db.CreateProductTxParams{
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		Discount:        req.Discount,
		AttributeValues: attributeValues(req.Attributes),
		Barcode:         req.Barcode,
		ColorID:         req.ColorID,
		SizeID:          req.SizeID,
		ColorIDs:        req.ColorIDs,
		SizeIDs:         req.SizeIDs,
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
	ID          int64      `json:"id" binding:"required,min=1"`
	Code        string     `json:"code" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active"`
	Attributes  Attributes `json:"attributes"`
}

// updateProduct edits the shared product-level fields (name/code/
// description/is_active) and the product's attributes — not any variant's
// own price/barcode/color/size, which is updateProductVariant's job below.
// The two are separate endpoints because they're separate rows now: a
// product can have several variants, and editing "the product" shouldn't
// require picking one of them.
func (server *Server) updateProduct(ctx *gin.Context) {
	var req updateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	result, err := server.store.UpdateProductTx(ctx, db.UpdateProductTxParams{
		ID:              req.ID,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		IsActive:        req.IsActive,
		AttributeValues: attributeValues(req.Attributes),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	ctx.JSON(http.StatusOK, gin.H{
		"product":    result.Product,
		"attributes": result.ProductAttributes,
	})
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

// variantResponse flattens db.ProductVariant so nullable columns serialise
// as plain JSON (null or a number) instead of sql.NullInt64's
// {"Int64":n,"Valid":b} shape the UI would otherwise have to unwrap.
type variantResponse struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	ColorID   *int64    `json:"color_id"`
	SizeID    *int64    `json:"size_id"`
	Barcode   string    `json:"barcode"`
	Price     *int64    `json:"price"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toVariantResponse(v db.ProductVariant) variantResponse {
	out := variantResponse{
		ID:        v.ID,
		ProductID: v.ProductID,
		Barcode:   v.Barcode,
		IsActive:  v.IsActive,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
	if v.ColorID.Valid {
		out.ColorID = &v.ColorID.Int64
	}
	if v.SizeID.Valid {
		out.SizeID = &v.SizeID.Int64
	}
	if v.Price.Valid {
		out.Price = &v.Price.Int64
	}
	return out
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

	variants, err := server.store.ListProductVariantsByProduct(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	variantsOut := make([]variantResponse, 0, len(variants))
	for _, v := range variants {
		variantsOut = append(variantsOut, toVariantResponse(v))
	}

	// TODO: use envelop
	ctx.JSON(http.StatusOK, gin.H{
		"product":    product,
		"attributes": attributes,
		"variants":   variantsOut,
	})
}

type listProductRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
	// Optional filters. Empty search / nil pointer / empty slice each mean
	// "don't filter on this". attribute_value_ids is an AND across values.
	Search           string     `form:"search"`
	IsActive         *bool      `form:"is_active"`
	HasVariants      *bool      `form:"has_variants"`
	CreatedFrom      *time.Time `form:"created_from" time_format:"2006-01-02"`
	CreatedTo        *time.Time `form:"created_to" time_format:"2006-01-02"`
	AttributeValueID []int64    `form:"attribute_value_ids"`
}

func nullBool(v *bool) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *v, Valid: true}
}

func nullTime(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}

func (server *Server) listProducts(ctx *gin.Context) {
	var req listProductRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	// created_to is a date; make it inclusive of that whole day.
	createdTo := req.CreatedTo
	if createdTo != nil {
		end := createdTo.AddDate(0, 0, 1)
		createdTo = &end
	}
	if req.AttributeValueID == nil {
		req.AttributeValueID = []int64{}
	}

	products, err := server.store.ListProducts(ctx, db.ListProductsParams{
		PageLimit:         req.PageSize,
		PageOffset:        req.PageID,
		Search:            req.Search,
		IsActive:          nullBool(req.IsActive),
		HasVariants:       nullBool(req.HasVariants),
		CreatedFrom:       nullTime(req.CreatedFrom),
		CreatedTo:         nullTime(createdTo),
		AttributeValueIds: req.AttributeValueID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountProducts(ctx, db.CountProductsParams{
		Search:            req.Search,
		IsActive:          nullBool(req.IsActive),
		HasVariants:       nullBool(req.HasVariants),
		CreatedFrom:       nullTime(req.CreatedFrom),
		CreatedTo:         nullTime(createdTo),
		AttributeValueIds: req.AttributeValueID,
	})
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
		// The product (or a variant of it, which cascade-deletes) is still
		// referenced by an invoice/purchase/transfer line — it can't go.
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "foreign_key_violation" {
			server.writeError(ctx, http.StatusConflict, errInUse)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

type createProductVariantRequest struct {
	ProductID int64  `json:"product_id" binding:"required,min=1"`
	ColorID   *int64 `json:"color_id"`
	SizeID    *int64 `json:"size_id"`
	Price     *int64 `json:"price"`
	Barcode   *int64 `json:"barcode"`
}

// createProductVariant adds one more variant to an existing product. The
// barcode is generated server-side unless one is supplied.
func (server *Server) createProductVariant(ctx *gin.Context) {
	var req createProductVariantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	variant, err := server.store.AddVariantTx(ctx, db.AddVariantTxParams{
		ProductID: req.ProductID,
		ColorID:   req.ColorID,
		SizeID:    req.SizeID,
		Price:     req.Price,
		Barcode:   req.Barcode,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				// same (product, color, size) or barcode already exists
				server.writeError(ctx, http.StatusConflict, err)
				return
			case "foreign_key_violation":
				server.writeError(ctx, http.StatusBadRequest, err)
				return
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	server.writeJSON(ctx, http.StatusOK, envelope{"variant": toVariantResponse(variant)})
}

type deleteProductVariantRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteProductVariant(ctx *gin.Context) {
	var req deleteProductVariantRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := server.store.DeleteProductVariant(ctx, req.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "foreign_key_violation" {
			server.writeError(ctx, http.StatusConflict, errInUse)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	server.branchHub.broadcastAll(branchWSMessage{Type: "products_updated"})
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
