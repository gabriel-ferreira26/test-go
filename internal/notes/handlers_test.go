package notes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *httptest.Server {
	h := NewHandlers(NewStore())
	return httptest.NewServer(h.Routes())
}

func TestHandlersCreateAndGet(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/notes", "application/json", strings.NewReader(`{"title":"Nota","content":"conteudo"}`))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	getResp, err := http.Get(srv.URL + "/notes/1")
	if err != nil {
		t.Fatalf("GET /notes/1 failed: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getResp.StatusCode)
	}
}

func TestHandlersCreateWithoutTitle(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/notes", "application/json", strings.NewReader(`{"content":"sem titulo"}`))
	if err != nil {
		t.Fatalf("POST /notes failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandlersGetNotFound(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/notes/999")
	if err != nil {
		t.Fatalf("GET /notes/999 failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandlersDelete(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	http.Post(srv.URL+"/notes", "application/json", strings.NewReader(`{"title":"Nota"}`))

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/notes/1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /notes/1 failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}
