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
		maxCalls := 10
		limit := ratelimit.Limit{
			Count:  maxCalls,
			Within: time.Second,
		}
		svr := ratelimit.NewServer(limit)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		for numberOfCalls := 1; numberOfCalls <= maxCalls; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}
	})

	t.Run("get / return error with 429 when Limit exceeded", func(t *testing.T) {
		const limitCalls = 10
		limit := ratelimit.Limit{
			Count:  limitCalls,
			Within: time.Second,
		}
		svr := ratelimit.NewServer(limit)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		recorder := httptest.NewRecorder()
		for numberOfCalls := 1; numberOfCalls <= limitCalls; numberOfCalls++ {
			svr.ServeHTTP(recorder, request)
		}

		recorder = httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusTooManyRequests)
		assertResponseBody(t, recorder, "error")
	})

	t.Run("different IP with separate limit", func(t *testing.T) {
		const clientIP1 = "IP1"
		const clientIP2 = "IP2"
		const clientIP3 = "IP3"

		maxCalls := 10
		limit := ratelimit.Limit{
			Count:  maxCalls,
			Within: time.Second,
		}
		svr := ratelimit.NewServer(limit)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = clientIP1
		for numberOfCalls := 1; numberOfCalls <= maxCalls; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}

		request.RemoteAddr = clientIP2
		for numberOfCalls := 1; numberOfCalls <= maxCalls; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}

		request.RemoteAddr = clientIP3
		for numberOfCalls := 1; numberOfCalls <= maxCalls; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}
	})

	t.Run("API available after limit period", func(t *testing.T) {
		const limitCalls = 10
		const limitPeriod = 500 * time.Millisecond
		limit := ratelimit.Limit{
			Count:  limitCalls,
			Within: limitPeriod,
		}
		svr := ratelimit.NewServer(limit)
		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		for numberOfCalls := 1; numberOfCalls <= limitCalls; numberOfCalls++ {
			svr.ServeHTTP(httptest.NewRecorder(), request)
		}

		time.Sleep(limitPeriod)
		time.Sleep(100 * time.Millisecond)

		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusOK)
		assertResponseBody(t, recorder, strconv.Itoa(1))
	})
}

func assertResponseBody(t *testing.T, recorder *httptest.ResponseRecorder, expected string) {
	got := recorder.Body.String()
	if got != expected {
		t.Errorf("expected response body is %q, got %q", expected, got)
	}
}

func assertResponseCode(t *testing.T, got, expected int) {
	t.Helper()
	if got != expected {
		t.Errorf("expect response status code %d, got %d", expected, got)
	}
}
