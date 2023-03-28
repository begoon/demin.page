package main

import (
	. "homepage/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestPingHandler(t *testing.T) {
	e := echo.New()
	req := T(http.NewRequest("GET", "/ping", nil))
	rec := httptest.NewRecorder()
	PingHandler(e.NewContext(req, rec))
	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", res.Status)
	}
	r := string(T(ioutil.ReadAll(res.Body)))
	if r != "pong" {
		t.Errorf("unexpected response from /ping %v", r)
	}
}

func TestPingRouteWithAPIKey(t *testing.T) {
	srv := httptest.NewServer(Configure())
	defer srv.Close()

	t.Run("with invalid key", func(t *testing.T) {
		req := T(http.NewRequest("GET", srv.URL+"/ping", nil))
		req.Header.Add("X-API-Key", "?")
		res := T(http.DefaultClient.Do(req))
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status Unauthorized, got %v", res.Status)
		}
	})

	t.Run("with valid key", func(t *testing.T) {
		req := T(http.NewRequest("GET", srv.URL+"/ping", nil))
		req.Header.Add("X-API-Key", "zorba")
		res := T(http.DefaultClient.Do(req))
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status OK, got %v", res.Status)
		}
	})
}

func TestPingRouteWithoutAPIKey(t *testing.T) {
	srv := httptest.NewServer(Configure())
	defer srv.Close()

	res := T(http.Get(srv.URL + "/ping"))
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status BadRequest, got %v", res.Status)
	}
}
