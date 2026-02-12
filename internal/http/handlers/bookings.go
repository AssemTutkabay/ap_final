package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/http/middleware"
	"ap_final/internal/service"
	"ap_final/internal/store"
)

type BookingsHandler struct {
	svc *service.BookingService

	masters store.MasterStore
	admins  store.SalonAdminStore
}

func NewBookingsHandler(
	svc *service.BookingService,
	masters store.MasterStore,
	admins store.SalonAdminStore,
) *BookingsHandler {
	return &BookingsHandler{svc: svc, masters: masters, admins: admins}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func isConflictErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "conflict:")
}

type UpdateBookingStatusRequest struct {
	Status       domain.BookingStatus `json:"status"`
	CancelReason string               `json:"cancelReason,omitempty"`
}

func (h *BookingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Создавать бронь может только client.
		if u.Role != domain.RoleClient {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}

		var req CreateBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad json")
			return
		}

		// ВАЖНО: clientId берем из токена, а не из body.
		b, err := h.svc.Create(r.Context(), u.ID, req.MasterID, req.ServiceID, req.StartAt)
		if err != nil {
			if isConflictErr(err) {
				writeErr(w, http.StatusConflict, err.Error())
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, b)

	case http.MethodGet:
		// RBAC фильтрация: кто что видит
		switch u.Role {
		case domain.RolePlatformAdmin:
			// может смотреть все, но может и фильтровать
			clientID := r.URL.Query().Get("clientId")
			masterID := r.URL.Query().Get("masterId")
			list, err := h.svc.List(r.Context(), clientID, masterID)
			if err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, list)

		case domain.RoleClient:
			// client видит только свои
			list, err := h.svc.List(r.Context(), u.ID, "")
			if err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, list)

		case domain.RoleMaster:
			// master видит только свои (через честную привязку user -> master)
			masterID, okm, err := h.masters.GetIDByUserID(r.Context(), u.ID)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			if !okm {
				writeErr(w, http.StatusForbidden, "forbidden: master is not linked to user")
				return
			}

			list, err := h.svc.List(r.Context(), "", masterID)
			if err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, list)

		case domain.RoleSalonAdmin:
			salonIDs, err := h.admins.ListSalonIDsForAdmin(r.Context(), u.ID)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			if len(salonIDs) == 0 {
				writeJSON(w, http.StatusOK, []domain.Booking{})
				return
			}

			list, err := h.svc.ListBySalonIDs(r.Context(), salonIDs)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			writeJSON(w, http.StatusOK, list)

		default:
			writeErr(w, http.StatusForbidden, "forbidden")
		}

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *BookingsHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/bookings/")
	if path == "" {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	// /bookings/{id}/status
	if strings.HasSuffix(path, "/status") {
		id := strings.TrimSuffix(path, "/status")
		id = strings.TrimSuffix(id, "/")
		if id == "" || strings.Contains(id, "/") {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		if r.Method != http.MethodPatch {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// статус менять могут platform_admin или salon_admin
		if u.Role != domain.RolePlatformAdmin && u.Role != domain.RoleSalonAdmin {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}

		// salon_admin: только если бронь относится к его салону
		if u.Role == domain.RoleSalonAdmin {
			bk, okb, err := h.svc.GetByID(r.Context(), id)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			if !okb {
				writeErr(w, http.StatusNotFound, "booking not found")
				return
			}
			allowed, err := h.admins.IsAdminOfSalon(r.Context(), u.ID, bk.SalonID)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			if !allowed {
				writeErr(w, http.StatusForbidden, "forbidden")
				return
			}
		}

		var req UpdateBookingStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad json")
			return
		}

		b, err := h.svc.UpdateStatus(r.Context(), id, req.Status, req.CancelReason)
		if err != nil {
			if isConflictErr(err) {
				writeErr(w, http.StatusConflict, err.Error())
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, b)
		return
	}

	// /bookings/{id}
	id := strings.TrimSuffix(path, "/")
	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// отменять может только client и только свою бронь
		if u.Role != domain.RoleClient {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}

		b, err := h.svc.CancelByClient(r.Context(), id, u.ID)
		if err != nil {
			if err.Error() == "forbidden" {
				writeErr(w, http.StatusForbidden, "forbidden")
				return
			}
			if err.Error() == "booking not found" {
				writeErr(w, http.StatusNotFound, "booking not found")
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, b)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
