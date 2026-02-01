package main

import (
	"fmt"
	"net/http"

	"ap_final/internal/router"
)

func main() {
	router.SetupRoutes()

	fmt.Println("Server running on http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}
