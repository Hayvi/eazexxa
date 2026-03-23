package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/betpro/server/internal/models"
	"github.com/betpro/server/internal/services"
	"github.com/betpro/server/pkg/logger"
)

type AuthHandler struct {
	authService  *services.AuthService
	userService  *services.UserService
	profileCache services.AuthProfileCache
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService, profileCache services.AuthProfileCache) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		userService:  userService,
		profileCache: profileCache,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string        `json:"token"`
	User  *models.User  `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := models.ValidateUsername(req.Username); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid username")
		return
	}

	if err := models.ValidateEmail(req.Email); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid email")
		return
	}

	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req.Username, req.Email, passwordHash)
	if err != nil {
		logger.Error("failed to create user", "error", err)
		respondError(w, http.StatusConflict, "User already exists")
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("failed to generate token", "error", err)
		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	user.Password = ""
	respondJSON(w, http.StatusCreated, AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !user.IsActive {
		respondError(w, http.StatusUnauthorized, "Account disabled")
		return
	}

	if err := h.authService.CheckPassword(req.Password, user.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Role)
	if err != nil {
		logger.Error("failed to generate token", "error", err)
		respondError(w, http.StatusInternalServerError, "Server error")
		return
	}

	profile := &models.Profile{
		ID:       user.ID,
		Balance:  user.Balance,
		IsActive: user.IsActive,
		Role:     user.Role,
	}
	_ = h.profileCache.Set(r.Context(), user.ID, profile)

	user.Password = ""
	respondJSON(w, http.StatusOK, AuthResponse{Token: token, User: user})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
