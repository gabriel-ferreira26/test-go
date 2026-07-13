package notes

import (
	"errors"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	s := NewStore()

	created := s.Create("Título", "Conteúdo")
	if created.ID != 1 {
		t.Fatalf("expected first note to have ID 1, got %d", created.ID)
	}

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got.Title != "Título" {
		t.Errorf("expected title %q, got %q", "Título", got.Title)
	}
}

func TestStoreGetNotFound(t *testing.T) {
	s := NewStore()

	_, err := s.Get(42)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	created := s.Create("Original", "Conteúdo original")

	updated, err := s.Update(created.ID, "Atualizado", "Conteúdo novo")
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

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	created := s.Create("Nota", "Conteúdo")

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("Delete returned unexpected error: %v", err)
	}

	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected note to be gone, got err %v", err)
	}
}

func TestStoreAllIsOrderedByID(t *testing.T) {
	s := NewStore()
	s.Create("Primeira", "")
	s.Create("Segunda", "")
	s.Create("Terceira", "")

	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 notes, got %d", len(all))
	}
	for i, n := range all {
		if n.ID != i+1 {
			t.Errorf("expected note at index %d to have ID %d, got %d", i, i+1, n.ID)
		}
	}
}
