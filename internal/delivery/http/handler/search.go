package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct{}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req domain.SearchQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid search query", err)
		return
	}

	mode := req.SearchMode
	if mode == "" {
		mode = "KEYWORD"
	}

	// Mocking search results layout
	response.Success(c, http.StatusOK, "Search completed via "+mode, []interface{}{})
}
