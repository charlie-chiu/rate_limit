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
	const dummyAddr = "dummyAddr"
	t.Run("whole process", func(t *testing.T) {
		// 200 ok, 429 too many req, wait, then 200 again
		const numOfRequests = 10
		const period = 2 * time.Second
		limit := ratelimit.Limit{
			Requests: numOfRequests,
			Within:   period,
		}
		svr := ratelimit.NewServer(limit)

		// 200 ok
		assertResponseOK(t, svr, numOfRequests, dummyAddr)

		// hitting limit, 429 too many req
		assertTooManyRequest(t, svr, dummyAddr)

		// within period, still 429 too many req
		time.Sleep(period / 2)
		assertTooManyRequest(t, svr, dummyAddr)

		// after period, 200 ok again
		time.Sleep(period)
		assertResponseOK(t, svr, numOfRequests, dummyAddr)
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

func assertTooManyRequest(t *testing.T, svr *ratelimit.Server, clientAddr string) {
	t.Helper()
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = clientAddr
	svr.ServeHTTP(recorder, request)
	assertResponseCode(t, recorder.Code, http.StatusTooManyRequests)
	assertResponseBody(t, recorder, "error")
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
