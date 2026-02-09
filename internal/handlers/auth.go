package handlers

import (
	"net/http"

	"jaycetrades.com/internal/schwab"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	schwabClient *schwab.Client
}

func NewAuthHandler(schwabClient *schwab.Client) *AuthHandler {
	return &AuthHandler{
		schwabClient: schwabClient,
	}
}

// Login redirects to Schwab OAuth
func (h *AuthHandler) Login(c *gin.Context) {
	authURL := h.schwabClient.GetAuthURL()
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback handles OAuth callback from Schwab
func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing authorization code",
		})
		return
	}

	if err := h.schwabClient.ExchangeCode(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to exchange code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully authenticated with Schwab",
		"status":  "ready",
	})
}

// Status returns current auth status
func (h *AuthHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"authenticated": h.schwabClient.IsAuthenticated(),
	})
}
