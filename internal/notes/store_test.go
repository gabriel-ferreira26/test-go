package notes

import (
	"errors"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	s := NewStore()

	created := s.Create(1, "Título", "Conteúdo")
	if created.ID != 1 {
		t.Fatalf("expected first note to have ID 1, got %d", created.ID)
	}

	got, err := s.Get(created.ID, 1)
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got.Title != "Título" {
		t.Errorf("expected title %q, got %q", "Título", got.Title)
	}
}

func TestStoreGetNotFound(t *testing.T) {
	s := NewStore()

	_, err := s.Get(42, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreGetWrongOwner(t *testing.T) {
	s := NewStore()
	created := s.Create(1, "Nota de Alice", "conteúdo")

	if _, err := s.Get(created.ID, 2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for a different owner, got %v", err)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	created := s.Create(1, "Original", "Conteúdo original")

	updated, err := s.Update(created.ID, 1, "Atualizado", "Conteúdo novo")
	if err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}
	if updated.Title != "Atualizado" {
		t.Errorf("expected title %q, got %q", "Atualizado", updated.Title)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) && updated.UpdatedAt != created.UpdatedAt {
		t.Errorf("expected UpdatedAt to change after update")
	}
}

func TestStoreUpdateWrongOwner(t *testing.T) {
	s := NewStore()
	created := s.Create(1, "Nota de Alice", "conteúdo")

	if _, err := s.Update(created.ID, 2, "Hackeado", "..."); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for a different owner, got %v", err)
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	created := s.Create(1, "Nota", "Conteúdo")

	if err := s.Delete(created.ID, 1); err != nil {
		t.Fatalf("Delete returned unexpected error: %v", err)
	}

	if _, err := s.Get(created.ID, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected note to be gone, got err %v", err)
	}
}

func TestStoreDeleteWrongOwner(t *testing.T) {
	s := NewStore()
	created := s.Create(1, "Nota de Alice", "conteúdo")

	if err := s.Delete(created.ID, 2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for a different owner, got %v", err)
	}
}

func TestStoreAllIsScopedToOwnerAndOrderedByID(t *testing.T) {
	s := NewStore()
	s.Create(1, "Primeira de Alice", "")
	s.Create(2, "Nota de Bob", "")
	s.Create(1, "Segunda de Alice", "")

	aliceNotes := s.All(1)
	if len(aliceNotes) != 2 {
		t.Fatalf("expected 2 notes for owner 1, got %d", len(aliceNotes))
	}
	if aliceNotes[0].ID > aliceNotes[1].ID {
		t.Errorf("expected notes ordered by ID, got %+v", aliceNotes)
	}

	bobNotes := s.All(2)
	if len(bobNotes) != 1 {
		t.Fatalf("expected 1 note for owner 2, got %d", len(bobNotes))
	}
}
