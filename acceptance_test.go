package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ratelimit"
)

func TestServer(t *testing.T) {
	t.Run("get / return number of calls", func(t *testing.T) {
		svr := ratelimit.NewServer()

		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusOK)
		assertResponseBody(t, recorder, strconv.Itoa(1))
	})

	t.Run("get / return error with 429 when Limit exceeded", func(t *testing.T) {
		svr := ratelimit.NewServer()

		request, _ := http.NewRequest(http.MethodGet, "/", nil)

		for numberOfCalls := 1; numberOfCalls <= ratelimit.Limit; numberOfCalls++ {
			recorder := httptest.NewRecorder()
			svr.ServeHTTP(recorder, request)
			expectedCalls := strconv.Itoa(numberOfCalls)
			assertResponseCode(t, recorder.Code, http.StatusOK)
			assertResponseBody(t, recorder, expectedCalls)
		}

		recorder := httptest.NewRecorder()
		svr.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusTooManyRequests)
		assertResponseBody(t, recorder, "error")
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
