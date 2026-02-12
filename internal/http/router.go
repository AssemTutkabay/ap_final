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

	// NEW: admin handlers
	AdminSalons   *handlers.AdminSalonsHandler
	AdminServices *handlers.AdminServicesHandler
	AdminMasters  *handlers.AdminMastersHandler

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

	// Private routes
	auth := deps.AuthMw.RequireAuth

	// /me - любая роль
	mux.Handle("/me", auth(http.HandlerFunc(deps.Me.Handle)))

	// /bookings
	mux.Handle("/bookings", auth(http.HandlerFunc(deps.Bookings.Handle)))
	mux.Handle("/bookings/", auth(http.HandlerFunc(deps.Bookings.HandleByID)))

	// POST /reviews - только client
	mux.Handle("/reviews",
		auth(middleware.RequireRoles(domain.RoleClient)(http.HandlerFunc(deps.Reviews.Handle))),
	)

	// -------------------------
	// NEW: Admin endpoints
	// -------------------------

	// Salons (platform_admin)
	mux.Handle("/admin/salons",
		auth(middleware.RequireRoles(domain.RolePlatformAdmin)(http.HandlerFunc(deps.AdminSalons.Handle))),
	)
	mux.Handle("/admin/salons/",
		auth(middleware.RequireRoles(domain.RolePlatformAdmin)(http.HandlerFunc(deps.AdminSalons.HandleByID))),
	)

	// Services (salon_admin for own salons + platform_admin)
	mux.Handle("/admin/services",
		auth(middleware.RequireRoles(domain.RoleSalonAdmin, domain.RolePlatformAdmin)(http.HandlerFunc(deps.AdminServices.Handle))),
	)
	mux.Handle("/admin/services/",
		auth(middleware.RequireRoles(domain.RoleSalonAdmin, domain.RolePlatformAdmin)(http.HandlerFunc(deps.AdminServices.HandleByID))),
	)

	// Masters (salon_admin for own salons + platform_admin)
	mux.Handle("/admin/masters",
		auth(middleware.RequireRoles(domain.RoleSalonAdmin, domain.RolePlatformAdmin)(http.HandlerFunc(deps.AdminMasters.Handle))),
	)
	mux.Handle("/admin/masters/",
		auth(middleware.RequireRoles(domain.RoleSalonAdmin, domain.RolePlatformAdmin)(http.HandlerFunc(deps.AdminMasters.HandleByID))),
	)

	return mux
}
