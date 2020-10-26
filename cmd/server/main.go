package main

import (
	"log"
	"net/http"
	"time"

	"ratelimit"
)

func main() {
	const addr = ":80"
	limit := ratelimit.Limit{
		Count:  60,
		Within: 60 * time.Second,
	}
	s := ratelimit.NewServer(limit)

	log.Fatal(http.ListenAndServe(addr, s))
}
