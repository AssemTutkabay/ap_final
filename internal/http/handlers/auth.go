package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ap_final/internal/domain"
	"ap_final/internal/http/middleware"
	"ap_final/internal/store"
)

type AuthHandler struct {
	users store.UserStore
}

func NewAuthHandler(users store.UserStore) *AuthHandler {
	return &AuthHandler{users: users}
}

type RegisterRequest struct {
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // игнорируем при self-register
}

type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.FullName = strings.TrimSpace(req.FullName)

	if req.Email == "" && req.Phone == "" {
		writeErr(w, http.StatusBadRequest, "email or phone required")
		return
	}
	if strings.TrimSpace(req.Password) == "" || len(req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "password must be at least 6 chars")
		return
	}

	// self-register всегда CLIENT. Роли админов/мастеров создаются через seed или админский эндпоинт.
	role := domain.RoleClient

	// check duplicates
	if req.Email != "" {
		_, ok, err := h.users.GetByEmail(r.Context(), req.Email)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if ok {
			writeErr(w, http.StatusConflict, "conflict: email already exists")
			return
		}
	}
	if req.Phone != "" {
		_, ok, err := h.users.GetByPhone(r.Context(), req.Phone)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if ok {
			writeErr(w, http.StatusConflict, "conflict: phone already exists")
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	u := domain.User{
		ID:           newID("u_"),
		Role:         role,
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		IsActive:     true,
	}

	created, err := h.users.Create(r.Context(), u)
	if err != nil {
		if err == store.ErrConflict {
			writeErr(w, http.StatusConflict, "conflict")
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	secret := []byte(os.Getenv("AUTH_SECRET"))
	if len(secret) == 0 {
		writeErr(w, http.StatusInternalServerError, "AUTH_SECRET not set")
		return
	}

	tok, err := middleware.MakeToken(secret, created.ID, time.Now())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{Token: tok, User: created})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Email == "" && req.Phone == "" {
		writeErr(w, http.StatusBadRequest, "email or phone required")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		writeErr(w, http.StatusBadRequest, "password required")
		return
	}

	var u domain.User
	var ok bool
	var err error

	if req.Email != "" {
		u, ok, err = h.users.GetByEmail(r.Context(), req.Email)
	} else {
		u, ok, err = h.users.GetByPhone(r.Context(), req.Phone)
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !ok || !u.IsActive {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secret := []byte(os.Getenv("AUTH_SECRET"))
	if len(secret) == 0 {
		writeErr(w, http.StatusInternalServerError, "AUTH_SECRET not set")
		return
	}

	tok, err := middleware.MakeToken(secret, u.ID, time.Now())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{Token: tok, User: u})
}

func newID(prefix string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return prefix + base64.RawURLEncoding.EncodeToString(b)
}
