package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
)

// errInUse is returned when a lookup row (color, size, attribute value) can't be
// deleted because a product still references it.
var errInUse = errors.New("still referenced by one or more products")

type createAttributeValueRequest struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

func (server *Server) createAttributeValue(ctx *gin.Context) {
	var req createAttributeValueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	attribute, err := server.store.GetAttribute(ctx, req.Attribute)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	arg := db.UpsertAttributeValueParams{
		AttributeID: attribute.ID,
		Value:       req.Value,
	}
	attributeValue, err := server.store.UpsertAttributeValue(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, attributeValue)
}

type createAttributeValuesRequest struct {
	Items []createAttributeValueRequest `json:"items" binding:"required,min=1"`
}

func (server *Server) createAttributeValues(ctx *gin.Context) {
	var req createAttributeValuesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	var results []db.AttributesValue
	for _, item := range req.Items {
		attribute, err := server.store.GetAttribute(ctx, item.Attribute)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}

		arg := db.UpsertAttributeValueParams{
			AttributeID: attribute.ID,
			Value:       item.Value,
		}
		attributeValue, err := server.store.UpsertAttributeValue(ctx, arg)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
		results = append(results, attributeValue)
	}
	ctx.JSON(http.StatusOK, results)
}

type getAttributeValue struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAttributeValue(ctx *gin.Context) {
	var req getAttributeValue
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	attributeValue, err := server.store.GetAttributeValue(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, attributeValue)
}

type listAttributeValuesRequest struct {
	AttributeName string `form:"attribute" binding:"required"`
	PageSize      int32  `form:"page_size,default=14" binding:"min=5,max=100"`
	PageID        int32  `form:"page_id,default=0" binding:"min=0"`
	Search        string `form:"search"`
	SortBy        string `form:"sort_by,default=created_at" binding:"oneof=id value created_at"`
	SortOrder     string `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

func (server *Server) listAttributeValues(ctx *gin.Context) {
	var req listAttributeValuesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	arg := db.ListAttributeValuesPageParams{
		AttributeName: req.AttributeName,
		Search:        req.Search,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
		PageSize:      req.PageSize,
		PageOffset:    req.PageID,
	}
	attributeValues, err := server.store.ListAttributeValuesPage(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	total, err := server.store.CountAttributeValues(ctx, db.CountAttributeValuesParams{
		AttributeName: req.AttributeName,
		Search:        req.Search,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"values": attributeValues, "total": total})
}

type updateAttributeValueRequest struct {
	ID    int64  `json:"id" binding:"required,min=1"`
	Value string `json:"value" binding:"required"`
}

func (server *Server) updateAttributeValue(ctx *gin.Context) {
	var req updateAttributeValueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.UpdateAttributeValueParams{
		ID:    req.ID,
		Value: req.Value,
	}
	attributeValue, err := server.store.UpdateAttributeValue(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			server.writeError(ctx, http.StatusConflict, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, attributeValue)
}

// listAttributes returns the fixed set of attribute types (category, brand, …)
// used to drive the management page's sidebar and the product form's fields.
func (server *Server) listAttributes(ctx *gin.Context) {
	attributes, err := server.store.ListAttributes(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"attributes": attributes})
}

// listAllAttributeValues returns every attribute value with its attribute name
// and creation time -- the management page groups these client-side.
func (server *Server) listAllAttributeValues(ctx *gin.Context) {
	values, err := server.store.ListAllAttributeValues(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"values": values})
}

type deleteAttributeValueRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAttributeValue(ctx *gin.Context) {
	var req deleteAttributeValueRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteAttributeValue(ctx, req.ID)
	if err != nil {
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
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
