package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Andrii-K-17/light-chat/internal/middleware"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/Andrii-K-17/light-chat/internal/response"
	"github.com/Andrii-K-17/light-chat/internal/services"
)

// AuthHandler manages HTTP authentication endpoints.
type AuthHandler struct {
	svc           *services.AuthService
	jwtSecret     string
	jwtExpiry     time.Duration
	refreshExpiry time.Duration
	cookieSecure  bool
}

// NewAuthHandler initializes and returns a new AuthHandler.
func NewAuthHandler(
	svc *services.AuthService,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshExpiry time.Duration,
	cookieSecure bool,
) *AuthHandler {
	return &AuthHandler{
		svc:           svc,
		jwtSecret:     jwtSecret,
		jwtExpiry:     jwtExpiry,
		refreshExpiry: refreshExpiry,
		cookieSecure:  cookieSecure,
	}
}

// issueTokenCookies sets both the access JWT and refresh token as HTTP-only cookies.
func (h *AuthHandler) issueTokenCookies(w http.ResponseWriter, pair *services.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    pair.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.jwtExpiry.Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.refreshExpiry.Seconds()),
	})
}

// clearTokenCookies removes both authentication cookies by expiring them.
func (h *AuthHandler) clearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth/refresh",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// registerRequest represents the registration payload.
type registerRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// Register creates a new user account and issues session cookies.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Email == "" || req.Username == "" || req.DisplayName == "" {
		response.Error(w, http.StatusUnprocessableEntity, "all fields are required")
		return
	}

	if len(req.Password) < 8 {
		response.Error(w, http.StatusUnprocessableEntity, "password must be at least 8 characters long")
		return
	}

	user, pair, err := h.svc.Register(
		req.Email,
		req.Username,
		req.DisplayName,
		req.Password,
		h.jwtSecret,
		h.jwtExpiry,
		h.refreshExpiry,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailTaken):
			response.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, services.ErrUsernameTaken):
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	h.issueTokenCookies(w, pair)
	response.JSON(w, http.StatusCreated, user)
}

// loginRequest represents the login payload.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user and provides session cookies.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, pair, err := h.svc.Login(
		strings.ToLower(strings.TrimSpace(req.Email)),
		req.Password,
		h.jwtSecret,
		h.jwtExpiry,
		h.refreshExpiry,
	)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.issueTokenCookies(w, pair)
	response.JSON(w, http.StatusOK, user)
}

// Refresh rotates the refresh token and issues a new token pair.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	pair, err := h.svc.Refresh(cookie.Value, h.jwtSecret, h.jwtExpiry, h.refreshExpiry)
	if err != nil {
		h.clearTokenCookies(w)
		if errors.Is(err, services.ErrRefreshTokenReused) {
			response.Error(w, http.StatusUnauthorized, "session revoked, please log in again")
			return
		}
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.issueTokenCookies(w, pair)
	response.JSON(w, http.StatusOK, map[string]string{"message": "refreshed"})
}

// Logout clears session cookies and invalidates the refresh token.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		_ = h.svc.Logout(cookie.Value)
	}
	h.clearTokenCookies(w)
	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me retrieves the currently authenticated user's profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.svc.GetByID(userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// updateProfileRequest represents the profile update payload.
type updateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Username    *string `json:"username"`
	Email       *string `json:"email"`
	Status      *string `json:"status"`
}

// UpdateProfile applies a partial patch to the authenticated user's profile.
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.svc.UpdateProfile(userID, repository.UserUpdateParams{
		DisplayName: req.DisplayName,
		Username:    req.Username,
		Email:       req.Email,
		Status:      req.Status,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, user)
}

// SearchUser finds a user by exact username match.
func (h *AuthHandler) SearchUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		response.Error(w, http.StatusBadRequest, "username parameter is required")
		return
	}

	user, err := h.svc.FindByUsername(username)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}
