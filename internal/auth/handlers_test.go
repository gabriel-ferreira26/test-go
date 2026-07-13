package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *httptest.Server {
	h := NewHandlers(NewStore())
	mux := http.NewServeMux()
	h.Routes(mux)
	return httptest.NewServer(mux)
}

func TestRegisterHandler(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(`{"username":"alice","password":"s3cret"}`))
	if err != nil {
		t.Fatalf("POST /auth/register failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestRegisterHandlerDuplicate(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body := `{"username":"alice","password":"s3cret"}`
	http.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(body))

	resp, err := http.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /auth/register failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func TestLoginHandler(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	http.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(`{"username":"alice","password":"s3cret"}`))

	resp, err := http.Post(srv.URL+"/auth/login", "application/json", strings.NewReader(`{"username":"alice","password":"s3cret"}`))
	if err != nil {
		t.Fatalf("POST /auth/login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Token == "" {
		t.Error("expected a non-empty token in the response")
	}
}

func TestLoginHandlerWrongPassword(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	http.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(`{"username":"alice","password":"s3cret"}`))

	resp, err := http.Post(srv.URL+"/auth/login", "application/json", strings.NewReader(`{"username":"alice","password":"errada"}`))
	if err != nil {
		t.Fatalf("POST /auth/login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}
