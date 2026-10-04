package handler

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

type OTPHandler struct {
	store sync.Map
}

func NewOTPHandler() *OTPHandler {
	return &OTPHandler{}
}

func (h *OTPHandler) RequestOTP(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		response.Error(c, http.StatusBadRequest, "phone is required", nil)
		return
	}

	val, _ := rand.Int(rand.Reader, big.NewInt(900000))
	otp := fmt.Sprintf("%06d", val.Int64()+100000)

	h.store.Store(phone, otp)

	// Simulate sending SMS via external provider
	fmt.Printf("SIMULATED SMS: Sending OTP %s to %s\n", otp, phone)

	go func() {
		time.Sleep(5 * time.Minute)
		h.store.Delete(phone) // expire OTP
	}()

	response.Success(c, http.StatusOK, "otp sent", nil)
}

func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	phone := c.Query("phone")
	otp := c.Query("otp")

	storedOTP, ok := h.store.Load(phone)
	if !ok || storedOTP.(string) != otp {
		response.Error(c, http.StatusUnauthorized, "invalid or expired otp", nil)
		return
	}

	h.store.Delete(phone)
	response.Success(c, http.StatusOK, "otp verified", nil)
}
