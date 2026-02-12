package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/http/middleware"
	"ap_final/internal/store"
)

type AdminMastersHandler struct {
	masters store.MasterStore
	admins  store.SalonAdminStore
}

func NewAdminMastersHandler(m store.MasterStore, a store.SalonAdminStore) *AdminMastersHandler {
	return &AdminMastersHandler{masters: m, admins: a}
}

type AdminCreateMasterRequest struct {
	ID           string `json:"id,omitempty"`
	SalonID      string `json:"salonId"`
	UserID       string `json:"userId,omitempty"` // опционально
	FullName     string `json:"fullName"`
	Bio          string `json:"bio,omitempty"`
	PortfolioURL string `json:"portfolioUrl,omitempty"`
	IsActive     bool   `json:"isActive"`
}

type AdminUpdateMasterRequest struct {
	FullName     string `json:"fullName"`
	Bio          string `json:"bio,omitempty"`
	PortfolioURL string `json:"portfolioUrl,omitempty"`
	IsActive     bool   `json:"isActive"`
}

// POST /admin/masters
func (h *AdminMastersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AdminCreateMasterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}

	req.ID = strings.TrimSpace(req.ID)
	req.SalonID = strings.TrimSpace(req.SalonID)
	req.UserID = strings.TrimSpace(req.UserID)
	req.FullName = strings.TrimSpace(req.FullName)
	req.Bio = strings.TrimSpace(req.Bio)
	req.PortfolioURL = strings.TrimSpace(req.PortfolioURL)

	if req.SalonID == "" || req.FullName == "" {
		writeErr(w, http.StatusBadRequest, "salonId and fullName required")
		return
	}
	if req.ID == "" {
		req.ID = newID("m_")
	}

	// salon_admin: только свои салоны
	if u.Role == domain.RoleSalonAdmin {
		allowed, err := h.admins.IsAdminOfSalon(r.Context(), u.ID, req.SalonID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !allowed {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	m := domain.Master{
		ID:           req.ID,
		SalonID:      req.SalonID,
		UserID:       req.UserID,
		FullName:     req.FullName,
		Bio:          req.Bio,
		PortfolioURL: req.PortfolioURL,
		IsActive:     req.IsActive,
	}

	created, err := h.masters.CreateForSalon(r.Context(), m)
	if err != nil {
		if err == store.ErrConflict {
			writeErr(w, http.StatusConflict, "conflict")
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// PUT/DELETE /admin/masters/{id}
func (h *AdminMastersHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/masters/")
	id = strings.TrimSuffix(id, "/")
	id = strings.TrimSpace(id)

	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	// salon_admin ownership: узнаем salonId через GetByID
	if u.Role == domain.RoleSalonAdmin {
		m, ok2, err := h.masters.GetByID(r.Context(), id)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !ok2 {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		allowed, err := h.admins.IsAdminOfSalon(r.Context(), u.ID, m.SalonID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !allowed {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	switch r.Method {
	case http.MethodPut:
		var req AdminUpdateMasterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad json")
			return
		}
		req.FullName = strings.TrimSpace(req.FullName)
		req.Bio = strings.TrimSpace(req.Bio)
		req.PortfolioURL = strings.TrimSpace(req.PortfolioURL)

		if req.FullName == "" {
			writeErr(w, http.StatusBadRequest, "fullName required")
			return
		}

		m := domain.Master{
			ID:           id,
			FullName:     req.FullName,
			Bio:          req.Bio,
			PortfolioURL: req.PortfolioURL,
			IsActive:     req.IsActive,
		}

		updated, err := h.masters.Update(r.Context(), m)
		if err != nil {
			if err == store.ErrNotFound {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			if err == store.ErrConflict {
				writeErr(w, http.StatusConflict, "conflict")
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		err := h.masters.Disable(r.Context(), id)
		if err != nil {
			if err == store.ErrNotFound {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
