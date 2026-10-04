package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type InstitutionalHandler struct {
	analytics *postgres.AnalyticsRepository
	graph     *postgres.GraphRepository
}

func NewInstitutionalHandler(analytics *postgres.AnalyticsRepository, graph *postgres.GraphRepository) *InstitutionalHandler {
	return &InstitutionalHandler{analytics: analytics, graph: graph}
}

const defaultAnalyticsWindow = 30 * 24 * time.Hour

// QueryGraph returns the institutional knowledge-graph subgraph around a
// search term, together with the edges connecting the matched nodes.
func (h *InstitutionalHandler) QueryGraph(c *gin.Context) {
	term := c.Query("query")
	limit := parsePositiveInt(c.Query("limit"), 25)

	if h.graph == nil {
		response.Success(c, http.StatusOK, "Graph queried successfully", gin.H{
			"nodes": []postgres.GraphNode{},
			"edges": []postgres.GraphEdge{},
		})
		return
	}

	nodes, err := h.graph.SearchNodes(c.Request.Context(), term, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GRAPH_QUERY_FAILED", err)
		return
	}

	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}

	edges, err := h.graph.EdgesForNodes(c.Request.Context(), ids)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GRAPH_QUERY_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, "Graph queried successfully", gin.H{
		"query":         term,
		"nodes":         nodes,
		"edges":         edges,
		"nodes_matched": len(nodes),
	})
}

// GetAnalytics reports event volume over the requested window, split by
// category and compared against the preceding window of the same length.
func (h *InstitutionalHandler) GetAnalytics(c *gin.Context) {
	window := defaultAnalyticsWindow
	if days := parsePositiveInt(c.Query("days"), 0); days > 0 {
		window = time.Duration(days) * 24 * time.Hour
	}

	if h.analytics == nil {
		response.Success(c, http.StatusOK, "Institutional analytics retrieved", gin.H{
			"window_days": int(window.Hours() / 24),
			"metrics":     []gin.H{},
		})
		return
	}

	ctx := c.Request.Context()
	now := time.Now()
	current, err := h.analytics.GetEventStats(ctx, now.Add(-window))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "ANALYTICS_FAILED", err)
		return
	}

	total := 0
	for _, stat := range current {
		total += stat.Count
	}

	previous, err := h.analytics.CountEvents(ctx, now.Add(-2*window))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "ANALYTICS_FAILED", err)
		return
	}

	response.Success(c, http.StatusOK, "Institutional analytics retrieved", gin.H{
		"window_days":     int(window.Hours() / 24),
		"total_events":    total,
		"previous_events": previous,
		"trend":           trendFor(total, previous),
		"by_category":     current,
	})
}

func trendFor(current, previous int) string {
	switch {
	case previous == 0 && current > 0:
		return "UP"
	case previous == 0:
		return "FLAT"
	case current > previous:
		return "UP"
	case current < previous:
		return "DOWN"
	default:
		return "FLAT"
	}
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
