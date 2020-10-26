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

func NewServer(l Limit) *Server {
	s := &Server{}
	limiter := newLimiter(l)

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(limiter.handle))
	s.Handler = router

	return s
}

type Limit struct {
	Count  int
	Within time.Duration
}

type limiter struct {
	calls       int
	period      time.Duration
	callCounter map[string]int
	mu          sync.Mutex
}

func newLimiter(limit Limit) *limiter {
	l := &limiter{
		calls:       limit.Count,
		period:      limit.Within,
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

	if numberOfCalls > l.calls {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, "error")
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, numberOfCalls)
	}
}

func (l *limiter) clearCounter() {
	for {
		time.Sleep(l.period)
		l.mu.Lock()
		l.callCounter = map[string]int{}
		l.mu.Unlock()
	}
}
