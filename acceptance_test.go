package ratelimit_test

import (
	"log"
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

		log.Println(svr)
		svr.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusOK)

		expected := strconv.Itoa(1)
		assertResponseBody(t, recorder, expected)
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
