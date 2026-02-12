package app

import (
	"database/sql"
	"net/http"

	httpx "ap_final/internal/http"
	"ap_final/internal/http/handlers"
	"ap_final/internal/http/middleware"
	"ap_final/internal/service"
	"ap_final/internal/store/postgres"
)

func Build(db *sql.DB) http.Handler {
	userStore := postgres.NewUserStore(db)
	salonStore := postgres.NewSalonStore(db)
	serviceStore := postgres.NewServiceStore(db)
	reviewStore := postgres.NewReviewStore(db)
	bookingStore := postgres.NewBookingStore(db)

	// NEW: stores for RBAC master + salon_admin
	masterStore := postgres.NewMasterStore(db)
	salonAdminStore := postgres.NewSalonAdminStore(db)

	bookingSvc := service.NewBookingService(bookingStore, serviceStore)

	authMw := middleware.NewAuthMiddleware(userStore)

	healthH := handlers.NewHealthHandler()
	salonsH := handlers.NewSalonsHandler(salonStore)
	servicesH := handlers.NewServicesHandler(serviceStore)

	// UPDATED: bookings handler now needs masterStore + salonAdminStore
	bookingsH := handlers.NewBookingsHandler(bookingSvc, masterStore, salonAdminStore)

	reviewsH := handlers.NewReviewsHandler(reviewStore, bookingStore)

	authH := handlers.NewAuthHandler(userStore)
	meH := handlers.NewMeHandler()

	return httpx.NewRouter(httpx.RouterDeps{
		Health:   healthH,
		Salons:   salonsH,
		Bookings: bookingsH,
		Services: servicesH,
		Reviews:  reviewsH,

		Auth:   authH,
		Me:     meH,
		AuthMw: authMw,
	})
}
