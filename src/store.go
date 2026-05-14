package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// store holds active sessions in memory. Sessions are ephemeral — they are
// not persisted and are lost on restart. User data lives in the DB only.
var store = &sessionStore{
	sessions: make(map[string]string),
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]string // token -> username
}

// CreateSession mints a cryptographically random token for the given username.
func (s *sessionStore) CreateSession(username string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("store: crypto/rand unavailable: " + err.Error())
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = username
	return token
}

// LookupSession resolves a token to a username.
func (s *sessionStore) LookupSession(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.sessions[token]
	return user, ok
}
