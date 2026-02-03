package http

import (
	"net/http"

	"ap_final/internal/http/handlers"
)

type RouterDeps struct {
	Health   *handlers.HealthHandler
	Salons   *handlers.SalonsHandler
	Bookings *handlers.BookingsHandler
	Services *handlers.ServicesHandler

}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", deps.Health.Handle)

	// Каталог
	mux.HandleFunc("/salons", deps.Salons.Handle)

	// Bookings (create + list)
	mux.HandleFunc("/bookings", deps.Bookings.Handle)

	// Bookings by id (cancel)
	mux.HandleFunc("/bookings/", deps.Bookings.HandleByID)

	// Services catalog
	mux.HandleFunc("/services", deps.Services.Handle)

	return mux
}
