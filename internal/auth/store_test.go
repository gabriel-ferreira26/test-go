package auth

import (
	"errors"
	"testing"
)

func TestRegisterAndLogin(t *testing.T) {
	s := NewStore()

	if _, err := s.Register("alice", "s3cret"); err != nil {
		t.Fatalf("Register returned unexpected error: %v", err)
	}

	token, err := s.Login("alice", "s3cret")
	if err != nil {
		t.Fatalf("Login returned unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty session token")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	s := NewStore()
	if _, err := s.Register("alice", "s3cret"); err != nil {
		t.Fatalf("first Register returned unexpected error: %v", err)
	}

	if _, err := s.Register("alice", "outrasenha"); !errors.Is(err, ErrUserExists) {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	s := NewStore()
	if _, err := s.Register("alice", "s3cret"); err != nil {
		t.Fatalf("Register returned unexpected error: %v", err)
	}

	if _, err := s.Login("alice", "senhaerrada"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUnknownUsername(t *testing.T) {
	s := NewStore()

	if _, err := s.Login("ninguem", "s3cret"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserIDForToken(t *testing.T) {
	s := NewStore()
	user, err := s.Register("alice", "s3cret")
	if err != nil {
		t.Fatalf("Register returned unexpected error: %v", err)
	}
	token, err := s.Login("alice", "s3cret")
	if err != nil {
		t.Fatalf("Login returned unexpected error: %v", err)
	}

	id, ok := s.UserIDForToken(token)
	if !ok {
		t.Fatal("expected token to resolve to a user ID")
	}
	if id != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, id)
	}

	if _, ok := s.UserIDForToken("token-invalido"); ok {
		t.Error("expected an unknown token to not resolve")
	}
}
