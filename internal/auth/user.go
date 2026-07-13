// Package auth handles user registration, login, and session verification.
package auth

// User represents a registered account.
//
// PasswordHash is never serialized to JSON on purpose: handlers build their
// own response maps instead of encoding User directly, so the hash never
// leaves the process.
type User struct {
	ID           int
	Username     string
	PasswordHash []byte
}
