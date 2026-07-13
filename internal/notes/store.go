package notes

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ErrNotFound is returned when a note with the given ID doesn't exist,
// or doesn't belong to the requesting owner.
var ErrNotFound = errors.New("note not found")

// Store is an in-memory, concurrency-safe collection of notes.
//
// Every method takes an ownerID so one user's notes stay invisible to
// everyone else. A real project would swap this for a database-backed
// implementation, but the handlers only depend on this struct's methods,
// so that swap wouldn't require touching the HTTP layer.
type Store struct {
	mu     sync.RWMutex
	notes  map[int]Note
	nextID int
}

// NewStore creates an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{
		notes:  make(map[int]Note),
		nextID: 1,
	}
}

// All returns every note owned by ownerID, ordered by ID.
func (s *Store) All(ownerID int) []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Note, 0)
	for _, n := range s.notes {
		if n.OwnerID == ownerID {
			result = append(result, n)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// Get returns a single note by ID, as long as it belongs to ownerID.
func (s *Store) Get(id, ownerID int) (Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notes[id]
	if !ok || n.OwnerID != ownerID {
		return Note{}, ErrNotFound
	}
	return n, nil
}

// Create adds a new note owned by ownerID and returns it with its
// assigned ID and timestamps.
func (s *Store) Create(ownerID int, title, content string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	n := Note{
		ID:        s.nextID,
		OwnerID:   ownerID,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.notes[n.ID] = n
	s.nextID++
	return n
}

// Update replaces the title and content of an existing note owned by ownerID.
func (s *Store) Update(id, ownerID int, title, content string) (Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, ok := s.notes[id]
	if !ok || n.OwnerID != ownerID {
		return Note{}, ErrNotFound
	}

	n.Title = title
	n.Content = content
	n.UpdatedAt = time.Now()
	s.notes[id] = n
	return n, nil
}

// Delete removes a note by ID, as long as it belongs to ownerID.
func (s *Store) Delete(id, ownerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, ok := s.notes[id]
	if !ok || n.OwnerID != ownerID {
		return ErrNotFound
	}
	delete(s.notes, id)
	return nil
}
