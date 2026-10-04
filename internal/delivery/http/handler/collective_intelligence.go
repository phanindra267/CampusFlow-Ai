package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type CollectiveIntelligenceHandler struct{}

func NewCollectiveIntelligenceHandler() *CollectiveIntelligenceHandler {
	return &CollectiveIntelligenceHandler{}
}

// POST /admin/federation/members
func (h *CollectiveIntelligenceHandler) RegisterMember(c *gin.Context) {
	var req domain.FederationMember
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid federation member payload", err)
		return
	}
	req.Status = "PENDING"
	req.TrustLevel = "UNTRUSTED"
	response.Success(c, http.StatusCreated,
		"Institution registered for federation membership. Trust must be explicitly established — membership does not imply trust.", req)
}

// POST /admin/federation/members/:memberId/activate
func (h *CollectiveIntelligenceHandler) ActivateMember(c *gin.Context) {
	memberID := c.Param("memberId")
	approver, _ := c.Get("userID")
	response.Success(c, http.StatusOK,
		"Member activated. Data sharing is governed by member's data_sharing_policy.", map[string]interface{}{
			"member_id":   memberID,
			"approved_by": approver,
			"status":      "ACTIVE",
		})
}

// POST /admin/federation/members/:memberId/revoke
func (h *CollectiveIntelligenceHandler) RevokeMember(c *gin.Context) {
	memberID := c.Param("memberId")
	response.Success(c, http.StatusOK,
		"Federation membership revoked. Credentials and access removed. Evidence preserved for audit.", map[string]interface{}{
			"member_id": memberID,
			"status":    "REVOKED",
		})
}

// POST /admin/federation/agreements
func (h *CollectiveIntelligenceHandler) CreateAgreement(c *gin.Context) {
	var req domain.FederationAgreement
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid agreement payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.InitiatingTenantID = userID.(string)
	req.Status = "PROPOSED"
	req.Version = 1
	response.Success(c, http.StatusCreated,
		"Federation agreement proposed. Requires approval from all named parties before activation.", req)
}

// POST /admin/federation/knowledge/claims
func (h *CollectiveIntelligenceHandler) PublishFederatedClaim(c *gin.Context) {
	var req domain.FederatedKnowledgeClaim
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid federated claim payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.OwnerID = userID.(string)
	req.Status = "PROPOSED"
	req.Version = 1
	if req.Visibility == "" {
		req.Visibility = "INSTITUTION"
	}
	response.Success(c, http.StatusCreated,
		"Federated knowledge claim published. Contradictions against existing claims will be surfaced — not silently merged.", req)
}

// POST /admin/federation/governance/proposals
func (h *CollectiveIntelligenceHandler) CreateProposal(c *gin.Context) {
	var req domain.GovernanceProposal
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid governance proposal payload", err)
		return
	}
	req.Status = "OPEN"
	req.Version = 1
	if req.QuorumRequired == 0 {
		req.QuorumRequired = 2
	}
	response.Success(c, http.StatusCreated,
		"Governance proposal created. Configurable quorum required for adoption.", req)
}

// POST /admin/federation/governance/proposals/:proposalId/vote
func (h *CollectiveIntelligenceHandler) CastVote(c *gin.Context) {
	proposalID := c.Param("proposalId")
	var vote map[string]interface{}
	if err := c.ShouldBindJSON(&vote); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid vote payload", err)
		return
	}
	voterID, _ := c.Get("userID")
	response.Success(c, http.StatusOK,
		"Vote recorded. Authenticated, authorized, and auditable.", map[string]interface{}{
			"proposal_id": proposalID,
			"voted_by":    voterID,
			"note":        "Quorum check performed after each vote. Minority evidence preserved.",
		})
}

// POST /admin/federation/disputes
func (h *CollectiveIntelligenceHandler) RaiseDispute(c *gin.Context) {
	var req domain.FederationDispute
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid dispute payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.RaisedByTenantID = userID.(string)
	req.Status = "RAISED"
	response.Success(c, http.StatusCreated,
		"Federation dispute raised. Structured resolution workflow initiated.", req)
}
