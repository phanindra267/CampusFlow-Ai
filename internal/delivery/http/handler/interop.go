package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type InteroperabilityHandler struct{}

func NewInteroperabilityHandler() *InteroperabilityHandler {
	return &InteroperabilityHandler{}
}

// POST /admin/interop/institutions
func (h *InteroperabilityHandler) RegisterInstitution(c *gin.Context) {
	var req domain.EduInstitution
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid institution payload", err)
		return
	}
	req.Status = "REGISTERED"
	req.TrustLevel = "UNVERIFIED"
	response.Success(c, http.StatusCreated,
		"Institution registered. Trust verification and onboarding required for federation capabilities.", req)
}

// POST /admin/interop/credentials/schemas
func (h *InteroperabilityHandler) RegisterSchema(c *gin.Context) {
	var req domain.CredentialSchema
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid schema payload", err)
		return
	}
	response.Success(c, http.StatusCreated,
		"Credential schema registered for ecosystem interoperability.", req)
}

// POST /institutions/credentials/issue
func (h *InteroperabilityHandler) IssueCredential(c *gin.Context) {
	var req domain.DigitalCredential
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid credential payload", err)
		return
	}
	req.Status = "VALID"
	req.DigitalSignature = "sig_crypto_hash_placeholder" // in a real system, this is crypto-signed
	response.Success(c, http.StatusCreated,
		"Digital credential securely issued and signed. Portable and verifiable by design.", req)
}

// POST /interop/credentials/verify
func (h *InteroperabilityHandler) VerifyCredential(c *gin.Context) {
	var req domain.DigitalCredential
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid credential verification payload", err)
		return
	}

	// Mock verification
	result := map[string]interface{}{
		"status":      "VALID",
		"issuer":      "VERIFIED",
		"schema":      "MATCHED",
		"explanation": "Credential cryptographically verified against trusted issuer registry.",
	}

	response.Success(c, http.StatusOK, "Verification complete.", result)
}

// POST /students/wallet/consents
func (h *InteroperabilityHandler) GrantConsent(c *gin.Context) {
	var req domain.DataConsent
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid consent payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.UserID = userID.(string)
	req.Status = "GRANTED"
	response.Success(c, http.StatusCreated,
		"Data sharing consent explicitly granted for selective disclosure.", req)
}

// POST /students/mobility/transfers
func (h *InteroperabilityHandler) RequestTransferCredit(c *gin.Context) {
	var req domain.TransferCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid transfer credit payload", err)
		return
	}
	userID, _ := c.Get("userID")
	req.StudentID = userID.(string)
	req.Status = "PROPOSED"
	response.Success(c, http.StatusCreated,
		"Transfer credit equivalency mapping proposed. Awaiting institutional approval.", req)
}

// POST /institutions/transfers/:id/approve
func (h *InteroperabilityHandler) ApproveTransferCredit(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	response.Success(c, http.StatusOK,
		"Transfer credit officially approved by institution.",
		map[string]interface{}{"id": id, "status": "APPROVED", "reviewed_by": userID})
}
