package handler

import (
	"context"
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventRepository interface {
	Create(ctx context.Context, e *domain.Event) error
	List(ctx context.Context) ([]domain.Event, error)
}

type EventHandler struct {
	repo EventRepository
}

func NewEventHandler(repo EventRepository) *EventHandler {
	return &EventHandler{repo: repo}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var e domain.Event
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}
	
	e.ID = uuid.New().String()
	userID, _ := c.Get("userID")
	e.OrganizerID = userID.(string)

	if err := h.repo.Create(c.Request.Context(), &e); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create event", err)
		return
	}
	response.Success(c, http.StatusCreated, "event created", e)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list events", err)
		return
	}
	response.Success(c, http.StatusOK, "events retrieved", events)
}
