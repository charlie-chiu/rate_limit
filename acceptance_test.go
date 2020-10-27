package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ratelimit"
)

func TestServer(t *testing.T) {
	t.Run("get / return number of calls", func(t *testing.T) {
		maxRequests := 10
		limit := ratelimit.Limit{
			Requests: maxRequests,
			Within:   time.Second,
		}
		svr := ratelimit.NewServer(limit)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		for numberOfCalls := 1; numberOfCalls <= maxRequests; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}
	})

	t.Run("return code 429 when Limit exceeded", func(t *testing.T) {
		const limitCalls = 4
		const limitWindow = 2 * time.Second
		limit := ratelimit.Limit{
			Requests: limitCalls,
			Within:   limitWindow,
		}
		svr := ratelimit.NewServer(limit)
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		for numberOfCalls := 1; numberOfCalls <= limitCalls; numberOfCalls++ {
			svr.ServeHTTP(httptest.NewRecorder(), request)
		}
		// sampling period is second
		time.Sleep(1200 * time.Millisecond)

		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusTooManyRequests)
		assertResponseBody(t, recorder, "error")
	})

	t.Run("different IP with separate limit", func(t *testing.T) {
		const clientIP1 = "IP1"
		const clientIP2 = "IP2"
		const clientIP3 = "IP3"

		maxRequests := 10
		limit := ratelimit.Limit{
			Requests: maxRequests,
			Within:   time.Second,
		}
		svr := ratelimit.NewServer(limit)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = clientIP1
		for numberOfCalls := 1; numberOfCalls <= maxRequests; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}

		request.RemoteAddr = clientIP2
		for numberOfCalls := 1; numberOfCalls <= maxRequests; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}

		request.RemoteAddr = clientIP3
		for numberOfCalls := 1; numberOfCalls <= maxRequests; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}
	})

	t.Run("API available after limit period", func(t *testing.T) {
		const maxRequests = 10
		const limitPeriod = time.Second
		limit := ratelimit.Limit{
			Requests: maxRequests,
			Within:   limitPeriod,
		}
		svr := ratelimit.NewServer(limit)
		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		for numberOfCalls := 1; numberOfCalls <= maxRequests; numberOfCalls++ {
			svr.ServeHTTP(httptest.NewRecorder(), request)
		}

		// wait for reset
		time.Sleep(limitPeriod)

		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusOK)
		assertResponseBody(t, recorder, strconv.Itoa(1))
	})
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
