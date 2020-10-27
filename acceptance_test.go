package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ratelimit"
)

func TestLimiter(t *testing.T) {
	t.Run("get / return number of calls", func(t *testing.T) {
		maxRequests := 10
		limit := ratelimit.Limit{
			Requests: maxRequests,
			Within:   time.Second,
		}
		svr := ratelimit.NewServer(limit)

		assertResponseOK(t, svr, maxRequests, "dummyAddr")
	})

	t.Run("return code 429 when Limit exceeded", func(t *testing.T) {
		const limitCalls = 4
		const limitWindow = 2 * time.Second
		limit := ratelimit.Limit{
			Requests: limitCalls,
			Within:   limitWindow,
		}
		svr := ratelimit.NewServer(limit)

		assertResponseOK(t, svr, limitCalls, "dummyAddr")
		// sampling period is second
		time.Sleep(1200 * time.Millisecond)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusTooManyRequests)
		assertResponseBody(t, recorder, "error")
	})

	t.Run("different IP with separate limit", func(t *testing.T) {

		maxRequests := 10
		limit := ratelimit.Limit{
			Requests: maxRequests,
			Within:   time.Second,
		}
		svr := ratelimit.NewServer(limit)

		assertResponseOK(t, svr, maxRequests, "addr1")
		assertResponseOK(t, svr, maxRequests, "addr2")
		assertResponseOK(t, svr, maxRequests, "addr3")
	})

	t.Run("API available after limit period", func(t *testing.T) {
		const maxRequests = 10
		const limitPeriod = time.Second
		limit := ratelimit.Limit{
			Requests: maxRequests,
			Within:   limitPeriod,
		}
		svr := ratelimit.NewServer(limit)

		assertResponseOK(t, svr, maxRequests, "dummyAddr")

		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		// wait for reset
		time.Sleep(limitPeriod)

		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusOK)
		assertResponseBody(t, recorder, strconv.Itoa(1))
	})
}

func assertResponseOK(t *testing.T, svr *ratelimit.Server, requests int, clientAddr string) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = clientAddr
	for numberOfRequests := 1; numberOfRequests <= requests; numberOfRequests++ {
		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusOK)
		assertResponseBody(t, recorder, strconv.Itoa(numberOfRequests))
	}
}

func assertResponseBody(t *testing.T, recorder *httptest.ResponseRecorder, expected string) {
	t.Helper()
	got := recorder.Body.String()
	if got != expected {
		t.Errorf("expected response body is %q, got %q", expected, got)
	}
}

func assertResponseCode(t *testing.T, got, expected int) {
	t.Helper()
	if got != expected {
		t.Errorf("expected response status code %d, got %d", expected, got)
	}
}
