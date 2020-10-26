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
	l := &limiter{
		callCounter: make(map[string]int),
	}

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(l.handle))
	s.Handler = router

	return s
}

const Limit = 60

type limiter struct {
	callCounter map[string]int
}

func (l *limiter) handle(w http.ResponseWriter, r *http.Request) {
	l.callCounter[r.RemoteAddr]++
	numberOfCalls := l.callCounter[r.RemoteAddr]

	if numberOfCalls > Limit {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, "error")
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, numberOfCalls)
	}

}
