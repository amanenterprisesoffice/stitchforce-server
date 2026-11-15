package v1

import (
	"net/http"
	"strconv"
	"strings"

	"stitchforce-server/config"
	"stitchforce-server/internal/controller/http/middleware"
	"stitchforce-server/internal/controller/http/v1/request"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/internal/usecase/input"
	"stitchforce-server/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	users    usecase.UserUseCase
	validate *validator.Validate
	appCfg   config.AppConfig
	jwt      *jwt.Manager
}

func NewUserHandler(
	users usecase.UserUseCase,
	appCfg config.AppConfig,
	validate *validator.Validate,
	jwtManager *jwt.Manager,
) *UserHandler {
	return &UserHandler{
		users:    users,
		validate: validate,
		appCfg:   appCfg,
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

	userAgent := c.Request.UserAgent()
	deviceName := c.GetHeader("X-Device-Name")

	if deviceName == "" {
		deviceName = "Unknown device"
	}

	u, session, err := h.users.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
		deviceName,
		c.ClientIP(),
		userAgent,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	token, err := h.jwt.Generate(
		u.ID,
		string(u.Role),
		u.OrganizationID,
		session.ID,
	)
	if err != nil {
		_ = h.users.RevokeSession(c.Request.Context(), session.ID)
		writeError(c, err)
		return
	}

	secure := h.appCfg.Env == "production"

	c.SetCookie(
		middleware.TokenKey,
		token,
		CookieMaxAge,
		"/",
		"",
		secure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"user":    u,
		"token":   token,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	sessionID := c.GetString(middleware.SessionIDKey)

	if sessionID != "" {
		if err := h.users.RevokeSession(c.Request.Context(), sessionID); err != nil {
			writeError(c, err)
			return
		}
	}

	middleware.RemoveCookie(c, h.appCfg.Env)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	var req request.VerifyEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.users.VerifyEmail(
		c.Request.Context(),
		req.Token,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
	})
}

func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.users.ForgotPassword(
		c.Request.Context(),
		req.Email,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If an account exists with that email, you will receive a password reset link.",
	})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.users.ResetPassword(
		c.Request.Context(),
		req.Token,
		req.NewPassword,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.UserID(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

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
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}

	u, err := h.users.GetByID(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) List(c *gin.Context) {
	organizationID := c.Param("organizationID")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "organization id is required",
		})
		return
	}

	const (
		Search = "search"
		Role   = "role"
		Limit  = "limit"
		Offset = "offset"
	)

	filter := input.UserFilter{
		Search: strings.TrimSpace(c.Query(Search)),
		Role:   strings.TrimSpace(c.Query(Role)),
		Limit:  20,
		Offset: 0,
	}

	if limit := c.Query(Limit); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil || value <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}
		filter.Limit = value
	}

	if offset := c.Query(Offset); offset != "" {
		value, err := strconv.Atoi(offset)
		if err != nil || value < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "offset must be a non-negative integer",
			})
			return
		}
		filter.Offset = value
	}

	users, err := h.users.List(
		c.Request.Context(),
		organizationID,
		filter,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := middleware.UserID(c)

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}

	var req request.UpdateUser
	var photo *input.UploadFile

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
			writeValidationError(c, err)
			return
		}

		file, header, err := c.Request.FormFile("photo")
		if err == nil {
			defer file.Close()

			photo = &input.UploadFile{
				Reader:      file,
				Size:        header.Size,
				ContentType: header.Header.Get("Content-Type"),
				Filename:    header.Filename,
			}
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid profile photo",
			})
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			writeValidationError(c, err)
			return
		}
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	updateInput := &input.UpdateUserInput{
		Name: req.Name,
		// Email: req.Email,
		Photo: photo,
	}

	if err := h.users.Update(
		c.Request.Context(),
		userID,
		updateInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}

	var req request.UpdateUser
	var photo *input.UploadFile

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
			writeValidationError(c, err)
			return
		}

		file, header, err := c.Request.FormFile("photo")
		if err == nil {
			defer file.Close()

			photo = &input.UploadFile{
				Reader:      file,
				Size:        header.Size,
				ContentType: header.Header.Get("Content-Type"),
				Filename:    header.Filename,
			}
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid profile photo",
			})
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			writeValidationError(c, err)
			return
		}
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(c, err)
		return
	}

	updateInput := &input.UpdateUserInput{
		Name: req.Name,
		// Email: req.Email,
		Role:  req.Role,
		Photo: photo,
	}

	if err := h.users.Update(
		c.Request.Context(),
		userID,
		updateInput,
	); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	currentUserID := middleware.UserID(c)
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}

	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot delete yourself",
		})
		return
	}

	if err := h.users.Delete(
		c.Request.Context(),
		userID,
	); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) DeleteMe(c *gin.Context) {
	userID := middleware.UserID(c)

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}

	var req request.DeleteMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := h.users.DeleteMe(
		c.Request.Context(),
		userID,
		req.Password,
	); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
