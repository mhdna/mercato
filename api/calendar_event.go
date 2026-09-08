package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// dateLayout ("2006-01-02") is declared in employee.go -- same package.

// listCalendarEvents returns every user-defined calendar event. The
// built-in retail calendar lives in the front end
// (ui/src/data/retailCalendarEvents.js) and is merged there.
func (server *Server) listCalendarEvents(ctx *gin.Context) {
	events, err := server.store.ListCalendarEvents(ctx)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"events": events})
}

type createCalendarEventRequest struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (server *Server) createCalendarEvent(ctx *gin.Context) {
	var req createCalendarEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		server.writeError(ctx, http.StatusBadRequest, errors.New("name is required"))
		return
	}

	start, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, errors.New("start_date must be YYYY-MM-DD"))
		return
	}
	end, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, errors.New("end_date must be YYYY-MM-DD"))
		return
	}
	if end.Before(start) {
		server.writeError(ctx, http.StatusBadRequest, errors.New("end_date must not be before start_date"))
		return
	}

	icon := strings.TrimSpace(req.Icon)
	if icon == "" {
		icon = "mdi-calendar-star"
	}
	color := strings.TrimSpace(req.Color)
	if color == "" {
		color = "primary"
	}

	event, err := server.store.CreateCalendarEvent(ctx, db.CreateCalendarEventParams{
		Name:      req.Name,
		Icon:      icon,
		Color:     color,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"event": event})
}

func (server *Server) deleteCalendarEvent(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id < 1 {
		server.writeError(ctx, http.StatusBadRequest, errors.New("invalid event id"))
		return
	}

	if err := server.store.DeleteCalendarEvent(ctx, id); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"deleted": true})
}
