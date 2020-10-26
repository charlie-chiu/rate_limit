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
	l := &limiter{}

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(l.handle))
	s.Handler = router

	return s
}

type limiter struct {
	callCounter int
}

func (l *limiter) handle(w http.ResponseWriter, r *http.Request) {
	l.callCounter++

	_, _ = fmt.Fprint(w, l.callCounter)
}
