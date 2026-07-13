package notes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabriel-ferreira26/test-go/internal/auth"
)

func decodeJSON(resp *http.Response, v any) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// newTestServer wires up the notes API behind auth, exactly like main.go
// does, and returns a ready-to-use client helper plus a valid bearer token
// for a freshly registered test user.
func newTestServer(t *testing.T) (srv *httptest.Server, token string) {
	t.Helper()

	authStore := auth.NewStore()
	authHandlers := auth.NewHandlers(authStore)

	notesStore := NewStore()
	notesHandlers := NewHandlers(notesStore)

	mux := http.NewServeMux()
	authHandlers.Routes(mux)
	notesHandlers.Routes(mux, authStore.RequireAuth)

	srv = httptest.NewServer(mux)

	if _, err := authStore.Register("alice", "s3cret"); err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}
	token, err := authStore.Login("alice", "s3cret")
	if err != nil {
		t.Fatalf("failed to log in test user: %v", err)
	}
	return srv, token
}

func authedRequest(t *testing.T, method, url, token, body string) *http.Response {
	t.Helper()

	var reqBody *strings.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	} else {
		reqBody = strings.NewReader("")
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestHandlersRequireAuth(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/notes")
	if err != nil {
		t.Fatalf("GET /notes failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestHandlersCreateAndGet(t *testing.T) {
	srv, token := newTestServer(t)
	defer srv.Close()

	resp := authedRequest(t, http.MethodPost, srv.URL+"/notes", token, `{"title":"Nota","content":"conteudo"}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	getResp := authedRequest(t, http.MethodGet, srv.URL+"/notes/1", token, "")
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getResp.StatusCode)
	}
}

func TestHandlersCreateWithoutTitle(t *testing.T) {
	srv, token := newTestServer(t)
	defer srv.Close()

	resp := authedRequest(t, http.MethodPost, srv.URL+"/notes", token, `{"content":"sem titulo"}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandlersGetNotFound(t *testing.T) {
	srv, token := newTestServer(t)
	defer srv.Close()

	resp := authedRequest(t, http.MethodGet, srv.URL+"/notes/999", token, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandlersDelete(t *testing.T) {
	srv, token := newTestServer(t)
	defer srv.Close()

	authedRequest(t, http.MethodPost, srv.URL+"/notes", token, `{"title":"Nota"}`).Body.Close()

	resp := authedRequest(t, http.MethodDelete, srv.URL+"/notes/1", token, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestHandlersNotesAreScopedPerUser(t *testing.T) {
	srv, aliceToken := newTestServer(t)
	defer srv.Close()

	authedRequest(t, http.MethodPost, srv.URL+"/notes", aliceToken, `{"title":"Nota da Alice"}`).Body.Close()

	authStoreClient := http.Client{}
	regResp, err := authStoreClient.Post(srv.URL+"/auth/register", "application/json", strings.NewReader(`{"username":"bob","password":"s3cret"}`))
	if err != nil {
		t.Fatalf("failed to register bob: %v", err)
	}
	regResp.Body.Close()

	loginResp, err := authStoreClient.Post(srv.URL+"/auth/login", "application/json", strings.NewReader(`{"username":"bob","password":"s3cret"}`))
	if err != nil {
		t.Fatalf("failed to log in bob: %v", err)
	}
	defer loginResp.Body.Close()

	var loginBody struct {
		Token string `json:"token"`
	}
	if err := decodeJSON(loginResp, &loginBody); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	// Bob should see an empty notebook, and shouldn't be able to reach Alice's note.
	listResp := authedRequest(t, http.MethodGet, srv.URL+"/notes", loginBody.Token, "")
	defer listResp.Body.Close()
	var notes []Note
	if err := decodeJSON(listResp, &notes); err != nil {
		t.Fatalf("failed to decode notes list: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected bob's notebook to be empty, got %d notes", len(notes))
	}

	getResp := authedRequest(t, http.MethodGet, srv.URL+"/notes/1", loginBody.Token, "")
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected bob to get 404 on alice's note, got %d", getResp.StatusCode)
	}
}
