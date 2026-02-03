package app

import (
	"ap_final/internal/domain"
	"net/http"

	httpx "ap_final/internal/http"
	"ap_final/internal/http/handlers"
	"ap_final/internal/service"
	"ap_final/internal/store/memory"
)

func Build() http.Handler {
	bookingStore := memory.NewBookingStore()

	servicesSeed := []domain.Service{
		{ID: "srv1", SalonID: "s1", Name: "Haircut", Price: 8000, DurationMin: 60},
		{ID: "srv2", SalonID: "s2", Name: "Manicure", Price: 6000, DurationMin: 45},
	}
	serviceStore := memory.NewServiceStore(servicesSeed)
	servicesH := handlers.NewServicesHandler(serviceStore)


	bookingSvc := service.NewBookingService(bookingStore, serviceStore)

	healthH := handlers.NewHealthHandler()
	salonsH := handlers.NewSalonsHandler()
	bookingsH := handlers.NewBookingsHandler(bookingSvc)

	return httpx.NewRouter(httpx.RouterDeps{
		Health:   healthH,
		Salons:   salonsH,
		Bookings: bookingsH,
		Services: servicesH,
	})
}
