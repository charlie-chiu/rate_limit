package ratelimit

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	http.Handler
}

func NewServer() *Server {
	s := &Server{}
	l := newLimiter()

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(l.handle))
	s.Handler = router

	return s
}

const Limit = 60

type limiter struct {
	callCounter map[string]int
	mu          sync.Mutex
}

func newLimiter() *limiter {
	l := &limiter{
		callCounter: make(map[string]int),
	}
	go l.clearCounter()
	return l
}

func (l *limiter) handle(w http.ResponseWriter, r *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
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

func (l *limiter) clearCounter() {
	for {
		time.Sleep(time.Second)
		l.mu.Lock()
		l.callCounter = map[string]int{}
		l.mu.Unlock()
	}
}
