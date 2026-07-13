package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// ErrUserExists is returned by Register when the username is already taken.
var ErrUserExists = errors.New("username already taken")

// ErrInvalidCredentials is returned by Login when the username or password is wrong.
var ErrInvalidCredentials = errors.New("invalid username or password")

// Store holds registered users and active session tokens in memory.
//
// Tokens here are opaque, random, server-side session identifiers rather
// than JWTs: the server is the only one that can tell whether a token is
// valid, which keeps the crypto surface small for a learning project. A
// production API would likely swap this for JWTs or a persisted session
// store, but the Handlers and middleware wouldn't need to change.
type Store struct {
	mu       sync.RWMutex
	users    map[string]User // keyed by username
	nextID   int
	sessions map[string]int // token -> user ID
}

// NewStore creates an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{
		users:    make(map[string]User),
		nextID:   1,
		sessions: make(map[string]int),
	}
}

// Register creates a new user with a bcrypt-hashed password.
func (s *Store) Register(username, password string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return User{}, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	u := User{ID: s.nextID, Username: username, PasswordHash: hash}
	s.users[username] = u
	s.nextID++
	return u, nil
}

// Login verifies credentials and, on success, issues a new session token.
func (s *Store) Login(username, password string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[username]
	if !ok {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := newToken()
	if err != nil {
		return "", err
	}
	s.sessions[token] = u.ID
	return token, nil
}

// UserIDForToken resolves an active session token to its user ID.
func (s *Store) UserIDForToken(token string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.sessions[token]
	return id, ok
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
