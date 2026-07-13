// Package notes contains the domain model and storage logic for the notes API.
package notes

import "time"

// Note represents a single note in the notebook, owned by one user.
type Note struct {
	ID        int       `json:"id"`
	OwnerID   int       `json:"owner_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
