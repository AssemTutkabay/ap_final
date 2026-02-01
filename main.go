package main

import (
	"log"
	"net/http"

	"ap_final/internal/app"
)

func main() {
	h := app.Build()

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", h); err != nil {
		log.Fatal(err)
	}
}
