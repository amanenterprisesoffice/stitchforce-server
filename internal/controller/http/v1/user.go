package v1

import (
	"net/http"
	"os"

	"stitchforce-server/internal/controller/http/middleware"
	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	users    usecase.UserUseCase
	validate *validator.Validate
	jwt      *jwt.Manager
}

func NewUserHandler(
	users usecase.UserUseCase,
	validate *validator.Validate,
	jwtManager *jwt.Manager,
) *UserHandler {
	return &UserHandler{
		users:    users,
		validate: validate,
		jwt:      jwtManager,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req request.Register

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	u, err := h.users.Register(
		c.Request.Context(),
		req.Name,
		req.Email,
		req.Password,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"user":    u,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req request.Login

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	u, err := h.users.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	token, err := h.jwt.Generate(u.ID, u.Role, u.OrganizationID)
	if err != nil {
		writeError(c, err)
		return
	}

	secure := os.Getenv("APP_ENV") == "production"

	// Set JWT in HTTP-only cookie
	c.SetCookie(
		middleware.TokenKey, // Cookie name
		token,               // JWT token
		3600,                // Max age: 1 hour
		"/",                 // Path
		"",                  // Domain (current domain)
		secure,              // Secure (HTTPS only)
		true,                // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"user":    u,
		"token":   token,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	secure := os.Getenv("APP_ENV") == "production"
	c.SetCookie(
		middleware.TokenKey,
		"",
		-1, // Expire immediately
		"/",
		"",
		secure, // Secure
		true,   // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.UserID(c)
	u, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User fetched successfully",
		"user":    u,
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	u, err := h.users.GetByID(
		c.Request.Context(), c.Param("id"),
	)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.users.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req request.UpdateUser

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	u := &entity.User{
		ID:    c.Param("id"),
		Name:  req.Name,
		Email: req.Email,
		Role:  req.Role,
	}

	if err := h.users.Update(c.Request.Context(), u); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.users.Delete(
		c.Request.Context(), c.Param("id"),
	); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
