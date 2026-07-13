package auth

import (
	"encoding/json"
	"errors"
	"net/http"

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

// Routes registers the auth endpoints onto mux.
func (h *Handlers) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) register(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if creds.Username == "" || creds.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, err := h.store.Register(creds.Username, creds.Password)
	if errors.Is(err, ErrUserExists) {
		httpx.WriteError(w, http.StatusConflict, "username already taken")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"id":       user.ID,
		"username": user.Username,
	})
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	token, err := h.store.Login(creds.Username, creds.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
