package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ratelimit"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
		log.Printf("Defaulting to port %s", port)
	}

	limit := ratelimit.Limit{
		Requests: 5,
		Within:   5 * time.Second,
	}
	s := ratelimit.NewServer(limit)

	log.Fatal(http.ListenAndServe(":"+port, s))
}
