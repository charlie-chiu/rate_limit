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

func NewServer(limit Limit) *Server {
	s := &Server{}
	l := newLimiter(limit)

	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(l.handle))
	s.Handler = router

	return s
}

type Limit struct {
	Requests int
	Within   time.Duration
}

type limiter struct {
	requests    int
	period      time.Duration
	callCounter map[string]int
	mu          sync.Mutex
}

func newLimiter(limit Limit) *limiter {
	l := &limiter{
		requests:    limit.Requests,
		period:      limit.Within,
		callCounter: make(map[string]int),
	}
	go l.resetPeriodically()
	return l
}

func (l *limiter) handle(w http.ResponseWriter, r *http.Request) {
	numberOfCalls, shouldHandle := l.shouldHandle(r)

	if shouldHandle {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, numberOfCalls)
	} else {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, "error")
	}
}

func (l *limiter) shouldHandle(r *http.Request) (numberOfRequest int, shouldHandle bool) {
	l.mu.Lock()
	l.callCounter[r.RemoteAddr]++
	numberOfCalls := l.callCounter[r.RemoteAddr]
	l.mu.Unlock()

	if numberOfCalls > l.requests {
		return numberOfCalls, false
	} else {
		return numberOfCalls, true
	}
}

func (l *limiter) resetPeriodically() {
	executeAt := time.Now().Add(l.period).Truncate(l.period)
	executeAfter := executeAt.Sub(time.Now())
	time.AfterFunc(executeAfter, func() {
		for {
			l.mu.Lock()
			l.callCounter = map[string]int{}
			l.mu.Unlock()

			time.Sleep(l.period)
		}
	})
}
