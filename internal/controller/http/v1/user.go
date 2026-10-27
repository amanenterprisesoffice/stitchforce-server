
package v1

import (
	"net/http"

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

	c.JSON(http.StatusCreated, u)
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

	token, err := h.jwt.Generate(u.ID, u.Role)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  u,
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
