package handlers

import (
	"net/http"

	"jaycetrades.com/internal/models"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	// In production, this would be a database or service
	users []models.User
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		users: []models.User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
		},
	}
}

func (h *UserHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, h.users)
}

func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	for _, user := range h.users {
		if string(rune(user.ID+'0')) == id {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		ID:    len(h.users) + 1,
		Name:  req.Name,
		Email: req.Email,
	}
	h.users = append(h.users, user)

	c.JSON(http.StatusCreated, user)
}
