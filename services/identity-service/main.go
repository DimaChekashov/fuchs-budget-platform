package main

import (
	"log"
	"net/http"

	"github.com/DimaChekashov/fuchs-budget-platform/services/identity-service/internal/router"
)

func main() {
	router := router.New()

	log.Println("identity-service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
