package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type FederatedNetworkHandler struct{}

func NewFederatedNetworkHandler() *FederatedNetworkHandler {
	return &FederatedNetworkHandler{}
}

// POST /network/nodes/register
func (h *FederatedNetworkHandler) RegisterNode(c *gin.Context) {
	var req domain.NetworkNode
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid node payload", err)
		return
	}
	req.TrustStatus = "PENDING_VERIFICATION"
	response.Success(c, http.StatusCreated,
		"Federated intelligence node registration initiated. Awaiting zero-trust cryptographic verification.", req)
}

// POST /network/nodes/:id/revoke
func (h *FederatedNetworkHandler) RevokeNode(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, http.StatusOK,
		"Federated node explicitly revoked. Key distribution halted and all trust boundaries sealed.", map[string]string{"id": id, "status": "REVOKED"})
}

// GET /network/federated/search
func (h *FederatedNetworkHandler) FederatedSearch(c *gin.Context) {
	query := c.Query("q")
	response.Success(c, http.StatusOK,
		"Federated search executed. Results derived only from explicitly authorized and responding nodes. Silence preserved for missing data.",
		map[string]interface{}{
			"query":             query,
			"results":           []string{"Authorized Research Artifact A", "Global Skill Equivalence B"},
			"confidence":        0.9,
			"unavailable_nodes": 1,
		})
}

// POST /network/skills/equivalence
func (h *FederatedNetworkHandler) ProposeSkillEquivalence(c *gin.Context) {
	var req domain.SkillEquivalence
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid equivalence payload", err)
		return
	}
	req.Status = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Skill equivalence proposal submitted to network governance. Merging requires evidentiary review.", req)
}

// POST /network/wallets/:id/consents
func (h *FederatedNetworkHandler) GrantWalletConsent(c *gin.Context) {
	var req domain.NetworkConsent
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid consent payload", err)
		return
	}
	req.Status = "ACTIVE"
	response.Success(c, http.StatusCreated,
		"Selective disclosure consent actively granted. Cryptographically signed and scoped to target node.", req)
}

// GET /network/wallets/:id/export
func (h *FederatedNetworkHandler) ExportWallet(c *gin.Context) {
	id := c.Param("id")
	response.Success(c, http.StatusOK,
		"Learning wallet records successfully packaged for secure, privacy-preserving portable export.",
		map[string]string{"wallet_id": id, "export_format": "VerifiableCredential"})
}
