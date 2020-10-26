package ratelimit

import (
	"fmt"
	"net/http"
)

type Server struct {
	http.Handler
}

func NewServer() *Server {
	s := &Server{}

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(handler))
	s.Handler = router

	return s
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, 1)
}
