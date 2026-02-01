package router

import (
	"net/http"

	"ap_final/internal/controller"
)

func SetupRoutes() {
	http.HandleFunc("/health", controller.Health)
}
