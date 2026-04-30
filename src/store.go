package main

import "sync"

// store holds the in-memory user credentials and active sessions.
// All access must go through the exported methods to ensure thread safety.
var store = &userStore{
	users:    make(map[string]string),
	sessions: make(map[string]string),
}

type userStore struct {
	mu       sync.Mutex
	users    map[string]string // username -> password
	sessions map[string]string // token -> username
}

// AddUser adds a new user. Returns false if the username is already taken.
func (s *userStore) AddUser(username, password string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[username]; exists {
		return false
	}
	s.users[username] = password
	return true
}

// Authenticate checks credentials. Returns the session token and true on success.
func (s *userStore) Authenticate(username, password string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pass, ok := s.users[username]; ok && pass == password {
		token := "tkn_" + username
		s.sessions[token] = username
		return token, true
	}
	return "", false
}

// LookupSession resolves a token to a username.
func (s *userStore) LookupSession(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.sessions[token]
	return user, ok
}

// Snapshot returns a copy of the users map (used for persistence).
func (s *userStore) Snapshot() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := make(map[string]string, len(s.users))
	for k, v := range s.users {
		snap[k] = v
	}
	return snap
}

// LoadUsers bulk-imports users (used on startup from disk or IndexedDB).
func (s *userStore) LoadUsers(imported map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range imported {
		s.users[k] = v
	}
}
