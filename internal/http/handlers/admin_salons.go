package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/store"
)

type AdminSalonsHandler struct {
	store store.SalonStore
}

func NewAdminSalonsHandler(s store.SalonStore) *AdminSalonsHandler {
	return &AdminSalonsHandler{store: s}
}

type AdminCreateSalonRequest struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	City               string `json:"city"`
	MarketplaceEnabled bool   `json:"marketplaceEnabled"`
}

type AdminUpdateSalonRequest struct {
	Name               string `json:"name"`
	City               string `json:"city"`
	MarketplaceEnabled bool   `json:"marketplaceEnabled"`
}

// POST /admin/salons
func (h *AdminSalonsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req AdminCreateSalonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}

	req.ID = strings.TrimSpace(req.ID)
	req.Name = strings.TrimSpace(req.Name)
	req.City = strings.TrimSpace(req.City)

	if req.Name == "" || req.City == "" {
		writeErr(w, http.StatusBadRequest, "name and city required")
		return
	}
	if req.ID == "" {
		req.ID = newID("s_")
	}

	sal := domain.Salon{
		ID:                 req.ID,
		Name:               req.Name,
		City:               req.City,
		MarketplaceEnabled: req.MarketplaceEnabled,
	}

	created, err := h.store.Create(r.Context(), sal)
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

// PUT/DELETE /admin/salons/{id}
func (h *AdminSalonsHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/salons/")
	id = strings.TrimSuffix(id, "/")
	id = strings.TrimSpace(id)

	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req AdminUpdateSalonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad json")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.City = strings.TrimSpace(req.City)
		if req.Name == "" || req.City == "" {
			writeErr(w, http.StatusBadRequest, "name and city required")
			return
		}

		sal := domain.Salon{
			ID:                 id,
			Name:               req.Name,
			City:               req.City,
			MarketplaceEnabled: req.MarketplaceEnabled,
		}

		updated, err := h.store.Update(r.Context(), sal)
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
		err := h.store.Delete(r.Context(), id)
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
