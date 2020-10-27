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
	callCounter map[string]map[int64]int
	mu          sync.Mutex
}

func newLimiter(limit Limit) *limiter {
	l := &limiter{
		requests:    limit.Requests,
		period:      limit.Within,
		callCounter: make(map[string]map[int64]int),
	}
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
	// increase request count
	now := time.Now().Unix()

	// first request from this addr
	if _, ok := l.callCounter[r.RemoteAddr]; !ok {
		l.callCounter[r.RemoteAddr] = map[int64]int{now: 1}
		return 1, true
	}

	l.callCounter[r.RemoteAddr][now]++
	// sum reqs in window
	windowStart := time.Now().Add(l.period*-1).Unix() + 1
	for s := windowStart; s <= now; s++ {
		numberOfRequest += l.callCounter[r.RemoteAddr][s]

		if numberOfRequest > l.requests {
			return 0, false
		}
	}

	return numberOfRequest, true
}
