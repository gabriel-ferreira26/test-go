package notes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gabriel-ferreira26/test-go/internal/auth"
	"github.com/gabriel-ferreira26/test-go/internal/httpx"
)

// Handlers wires HTTP requests to the Store.
type Handlers struct {
	store *Store
}

// NewHandlers creates Handlers backed by the given Store.
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// Routes registers the notes endpoints onto mux. Every route is wrapped
// with protect (typically an auth.Store's RequireAuth), so a caller must
// be authenticated before reaching any handler here.
func (h *Handlers) Routes(mux *http.ServeMux, protect func(http.Handler) http.Handler) {
	mux.Handle("GET /notes", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /notes", protect(http.HandlerFunc(h.create)))
	mux.Handle("GET /notes/{id}", protect(http.HandlerFunc(h.get)))
	mux.Handle("PUT /notes/{id}", protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /notes/{id}", protect(http.HandlerFunc(h.delete)))
}

type noteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := auth.UserIDFromContext(r.Context())
	httpx.WriteJSON(w, http.StatusOK, h.store.All(ownerID))
}

func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := auth.UserIDFromContext(r.Context())

	var in noteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.Title == "" {
		httpx.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}

	note := h.store.Create(ownerID, in.Title, in.Content)
	httpx.WriteJSON(w, http.StatusCreated, note)
}

func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := auth.UserIDFromContext(r.Context())

	id, err := idFromPath(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	note, err := h.store.Get(id, ownerID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "note not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, note)
}

func (h *Handlers) update(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := auth.UserIDFromContext(r.Context())

	id, err := idFromPath(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	var in noteInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.Title == "" {
		httpx.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}

	note, err := h.store.Update(id, ownerID, in.Title, in.Content)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "note not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, note)
}

func (h *Handlers) delete(w http.ResponseWriter, r *http.Request) {
	ownerID, _ := auth.UserIDFromContext(r.Context())

	id, err := idFromPath(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	if err := h.store.Delete(id, ownerID); errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "note not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func idFromPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}
