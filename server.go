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
	requests   int
	period     time.Duration
	reqCounter map[string]map[int64]int
	mu         sync.Mutex
}

func newLimiter(limit Limit) *limiter {
	l := &limiter{
		requests:   limit.Requests,
		period:     limit.Within,
		reqCounter: make(map[string]map[int64]int),
	}
	return l
}

func (l *limiter) handle(w http.ResponseWriter, r *http.Request) {
	shouldHandle, numberOfReqs := l.shouldHandle(r)

	if shouldHandle {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, numberOfReqs)
	} else {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, "error")
	}
}

func (l *limiter) shouldHandle(r *http.Request) (shouldHandle bool, numberOfReqs int) {
	userIP := readUserIP(r)

	// increase request count
	now := time.Now().Unix()

	l.mu.Lock()
	defer l.mu.Unlock()
	// first request from this addr
	if _, ok := l.reqCounter[userIP]; !ok {
		l.reqCounter[userIP] = map[int64]int{now: 1}
		return true, 1
	}

	l.reqCounter[userIP][now]++
	// sum reqs in window
	windowStart := time.Now().Add(l.period*-1).Unix() + 1
	for s := windowStart; s <= now; s++ {
		numberOfReqs += l.reqCounter[userIP][s]

		if numberOfReqs > l.requests {
			return false, 0
		}
	}

	return true, numberOfReqs
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
