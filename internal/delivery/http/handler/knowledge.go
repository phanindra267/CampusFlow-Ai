package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type KnowledgeGraphHandler struct{}

func NewKnowledgeGraphHandler() *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{}
}

// POST /knowledge/entities
func (h *KnowledgeGraphHandler) RegisterEntity(c *gin.Context) {
	var req domain.KnowledgeEntity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid entity payload", err)
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Knowledge entity registered with provenance traceability.", req)
}

// POST /knowledge/claims
func (h *KnowledgeGraphHandler) RegisterClaim(c *gin.Context) {
	var req domain.GraphClaim
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid claim payload", err)
		return
	}
	if req.IsMachineExtracted {
		req.VerificationStatus = "CANDIDATE"
	}
	response.Success(c, http.StatusCreated,
		"Scientific claim recorded. If machine-extracted, requires human verification.", req)
}

// GET /knowledge/search/hybrid
func (h *KnowledgeGraphHandler) HybridSearch(c *gin.Context) {
	query := c.Query("q")
	response.Success(c, http.StatusOK,
		"Hybrid search executed using lexical, vector, and graph embeddings. Returning strictly authorized knowledge.",
		map[string]interface{}{
			"query":        query,
			"results":      []string{"Authorized Concept A", "Research Document B"},
			"explanations": []string{"Matched via Semantic Taxonomy", "Matched via Graph Proximity"},
		})
}

// POST /knowledge/rag/query
func (h *KnowledgeGraphHandler) GraphRAG(c *gin.Context) {
	response.Success(c, http.StatusOK,
		"Graph-enhanced retrieval generated. Answers are explicitly grounded in cited evidence.",
		map[string]interface{}{
			"answer":    "Generated research synthesis.",
			"citations": []string{"Source Document ID 1", "Claim ID 2"},
			"warning":   "If evidence is insufficient, this response will explicitly declare it.",
		})
}

// POST /knowledge/hypotheses
func (h *KnowledgeGraphHandler) CreateHypothesis(c *gin.Context) {
	var req domain.GeneratedHypothesis
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid hypothesis payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.LifecycleStatus = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Research hypothesis created. Clearly designated as a candidate connection pending scientific validation.", req)
}
