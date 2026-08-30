package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// listPageQuery is the shared query-string shape for the "search + paginate
// (+ sort)" list endpoints — coupons, suppliers, price lists, discount
// lists, currencies, loans, transfers, expenses, purchases, products and
// their list-item sub-resources. PageID is a row offset (it maps straight
// to ListXParams.PageOffset), not a 1-based page number.
//
// Handlers with deliberately different bounds keep their own request
// structs: asset thumbnails cap PageSize at 10, inventory stock allows 500,
// and colors/sizes treat an absent PageSize as "return everything" for the
// picker dropdowns.
type listPageQuery struct {
	PageSize  int32  `form:"page_size,default=10" binding:"min=5,max=100"`
	PageID    int32  `form:"page_id,default=0" binding:"min=0"`
	Search    string `form:"search"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// bindListPageQuery parses the shared pagination params, writing a 400 and
// returning ok=false when they are malformed.
func (server *Server) bindListPageQuery(ctx *gin.Context) (listPageQuery, bool) {
	var q listPageQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return q, false
	}
	return q, true
}

// respondList runs the row query then the count query and renders the
// standard paginated body — {"<key>": rows, "total": n} — turning either
// failure into a 500. It replaces the identical fetch / count / marshal
// tail that every list handler used to spell out by hand.
func respondList[T any](
	server *Server, ctx *gin.Context, key string,
	list func() ([]T, error),
	count func() (int64, error),
) {
	rows, err := list()
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	total, err := count()
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{key: rows, "total": total})
}
