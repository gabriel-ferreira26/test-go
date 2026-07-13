package notes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// Handlers wires HTTP requests to the Store.
type Handlers struct {
	store *Store
}

// NewHandlers creates Handlers backed by the given Store.
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// Routes builds the HTTP router for the notes API.
//
// It uses the Go 1.22+ ServeMux, which supports method matching
// (e.g. "GET /notes") and path wildcards (e.g. "{id}") without
// needing a third-party router.
func (h *Handlers) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes", h.list)
	mux.HandleFunc("POST /notes", h.create)
	mux.HandleFunc("GET /notes/{id}", h.get)
	mux.HandleFunc("PUT /notes/{id}", h.update)
	mux.HandleFunc("DELETE /notes/{id}", h.delete)
	return mux
}

type noteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.All())
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	var in noteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	note := h.store.Create(in.Title, in.Content)
	writeJSON(w, http.StatusCreated, note)
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	note, err := h.store.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *Handlers) update(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	var in noteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	note, err := h.store.Update(id, in.Title, in.Content)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *Handlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	if err := h.store.Delete(id); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func idFromPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
