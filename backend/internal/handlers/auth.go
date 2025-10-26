package handlers

import (
	"net/http"

	"visekai/backend/internal/middleware"
	"visekai/backend/internal/models"
	"visekai/backend/internal/repository"
	"visekai/backend/internal/services"
	"visekai/backend/pkg/validator"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *services.AuthService
	userRepo    *repository.UserRepository
	validator   *validator.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
		validator:   validator.New(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegistration

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			"Invalid request body",
			nil,
		))
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			err.Error(),
			nil,
		))
		return
	}

	// Validate password strength
	if err := validator.ValidatePassword(req.Password, validator.DefaultPasswordStrength()); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_002",
			err.Error(),
			nil,
		))
		return
	}

	// Register user
	_, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"AUTH_001",
			err.Error(),
			nil,
		))
		return
	}

	// Generate tokens
	loginReq := models.UserLogin{
		Email:    req.Email,
		Password: req.Password,
	}
	authResponse, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"SYS_001",
			"Failed to generate tokens",
			nil,
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(
		authResponse,
		"User registered successfully",
	))
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLogin

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			"Invalid request body",
			nil,
		))
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			err.Error(),
			nil,
		))
		return
	}

	// Login user
	authResponse, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_001",
			err.Error(),
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		authResponse,
		"Login successful",
	))
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is typically handled client-side
	// by removing the token. Here we can add token to a blacklist if needed.
	c.JSON(http.StatusOK, models.NewSuccessResponse(
		nil,
		"Logout successful",
	))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"VAL_001",
			"Refresh token is required",
			nil,
		))
		return
	}

	// Refresh tokens
	authResponse, err := h.authService.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_002",
			"Invalid or expired refresh token",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		authResponse,
		"Tokens refreshed successfully",
	))
}

// GetCurrentUser returns the currently authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"AUTH_003",
			err.Error(),
			nil,
		))
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"RES_001",
			"User not found",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(
		user.ToResponse(),
		"User retrieved successfully",
	))
}
