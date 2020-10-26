package main

import (
	"log"
	"net/http"

	"ratelimit"
)

func main() {
	const addr = ":80"
	s := ratelimit.NewServer()

	log.Fatal(http.ListenAndServe(addr, s))
}
