package http

import (
	"net/http"

	"ap_final/internal/domain"
	"ap_final/internal/http/handlers"
	"ap_final/internal/http/middleware"
)

type RouterDeps struct {
	Health   *handlers.HealthHandler
	Salons   *handlers.SalonsHandler
	Bookings *handlers.BookingsHandler
	Services *handlers.ServicesHandler
	Reviews  *handlers.ReviewsHandler

	Auth   *handlers.AuthHandler
	Me     *handlers.MeHandler
	AuthMw *middleware.AuthMiddleware
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", deps.Health.Handle)

	// Public catalog
	mux.HandleFunc("/salons", deps.Salons.Handle)
	mux.HandleFunc("/salons/", deps.Reviews.HandleBySalon) // GET /salons/{id}/reviews - public
	mux.HandleFunc("/services", deps.Services.Handle)      // public (catalog by salonId)

	// Auth
	mux.HandleFunc("/auth/register", deps.Auth.Register)
	mux.HandleFunc("/auth/login", deps.Auth.Login)

	// Private routes (middleware must be non-nil in app.Build)
	auth := deps.AuthMw.RequireAuth

	// /me - любая роль
	mux.Handle("/me", auth(http.HandlerFunc(deps.Me.Handle)))

	// /bookings - только авторизованные (дальше внутри handler решим, что кому показывать)
	mux.Handle("/bookings", auth(http.HandlerFunc(deps.Bookings.Handle)))
	mux.Handle("/bookings/", auth(http.HandlerFunc(deps.Bookings.HandleByID)))

	// POST /reviews - только client (verified review)
	mux.Handle("/reviews",
		auth(middleware.RequireRoles(domain.RoleClient)(http.HandlerFunc(deps.Reviews.Handle))),
	)

	return mux
}
