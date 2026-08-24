package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

type createAttributeValueRequest struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

func (server *Server) createAttributeValue(ctx *gin.Context) {
	var req createAttributeValueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	attribute, err := server.store.GetAttribute(ctx, req.Attribute)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.UpsertAttributeValueParams{
		AttributeID: attribute.ID,
		Value:       req.Value,
	}
	attributeValue, err := server.store.UpsertAttributeValue(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var results []db.AttributesValue
	for _, item := range req.Items {
		attribute, err := server.store.GetAttribute(ctx, item.Attribute)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		arg := db.UpsertAttributeValueParams{
			AttributeID: attribute.ID,
			Value:       item.Value,
		}
		attributeValue, err := server.store.UpsertAttributeValue(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	attributeValue, err := server.store.GetAttributeValue(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, attributeValue)
}

type listAttributeValuesRequest struct {
	// AttributeID int64 `form:"attribute_id" binding:"min=1"`
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listAttributeValues(ctx *gin.Context) {
	var req listAttributeValuesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListAttributeValuesParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	attributeValues, err := server.store.ListAttributeValues(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, attributeValues)
}

type updateAttributeValueRequest struct {
	ID    int64  `json:"id" binding:"required,min=1"`
	Value string `json:"value" binding:"required"`
}

func (server *Server) updateAttributeValue(ctx *gin.Context) {
	var req updateAttributeValueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateAttributeValueParams{
		ID:    req.ID,
		Value: req.Value,
	}
	attributeValue, err := server.store.UpdateAttributeValue(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, attributeValue)
}
