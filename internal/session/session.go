// Package session provides session management for MCP connections.
package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Common errors.
var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionLimitReached = errors.New("maximum sessions reached")
)

// ClientInfo contains information about the connected client.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	IP      string `json:"ip,omitempty"`
}

// Session represents an MCP session.
type Session struct {
	ID           string         `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	LastActivity time.Time      `json:"last_activity"`
	ClientInfo   ClientInfo     `json:"client_info"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Initialized  bool           `json:"initialized"`
}

// Store defines the interface for session storage.
type Store interface {
	// Create creates a new session and returns it.
	Create(ctx context.Context, clientInfo ClientInfo) (*Session, error)

	// Get retrieves a session by ID.
	Get(ctx context.Context, id string) (*Session, error)

	// Touch updates the LastActivity timestamp.
	Touch(ctx context.Context, id string) error

	// Delete removes a session.
	Delete(ctx context.Context, id string) error

	// MarkInitialized marks a session as fully initialized.
	MarkInitialized(ctx context.Context, id string) error

	// Cleanup removes expired sessions and returns count removed.
	Cleanup(ctx context.Context) (int, error)

	// Count returns the number of active sessions.
	Count(ctx context.Context) (int, error)
}

// MemoryStore is an in-memory session store.
type MemoryStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	timeout  time.Duration
	maxSize  int
}

// MemoryStoreOptions configures the memory store.
type MemoryStoreOptions struct {
	Timeout time.Duration
	MaxSize int
}

// DefaultMemoryStoreOptions returns default options.
func DefaultMemoryStoreOptions() MemoryStoreOptions {
	return MemoryStoreOptions{
		Timeout: 10 * time.Minute,
		MaxSize: 10000,
	}
}

// NewMemoryStore creates a new in-memory session store.
func NewMemoryStore(opts MemoryStoreOptions) *MemoryStore {
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Minute
	}
	if opts.MaxSize == 0 {
		opts.MaxSize = 10000
	}

	store := &MemoryStore{
		sessions: make(map[string]*Session),
		timeout:  opts.Timeout,
		maxSize:  opts.MaxSize,
	}

	return store
}

// Create creates a new session.
func (s *MemoryStore) Create(ctx context.Context, clientInfo ClientInfo) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.sessions) >= s.maxSize {
		return nil, ErrSessionLimitReached
	}

	now := time.Now()
	session := &Session{
		ID:           uuid.New().String(),
		CreatedAt:    now,
		LastActivity: now,
		ClientInfo:   clientInfo,
		Metadata:     make(map[string]any),
		Initialized:  false,
	}

	s.sessions[session.ID] = session
	return session, nil
}

// Get retrieves a session by ID.
func (s *MemoryStore) Get(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if time.Since(session.LastActivity) > s.timeout {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// Touch updates the LastActivity timestamp.
func (s *MemoryStore) Touch(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}

	session.LastActivity = time.Now()
	return nil
}

// Delete removes a session.
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[id]; !ok {
		return ErrSessionNotFound
	}

	delete(s.sessions, id)
	return nil
}

// MarkInitialized marks a session as fully initialized.
func (s *MemoryStore) MarkInitialized(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}

	session.Initialized = true
	return nil
}

// Cleanup removes expired sessions.
func (s *MemoryStore) Cleanup(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	now := time.Now()

	for id, session := range s.sessions {
		if now.Sub(session.LastActivity) > s.timeout {
			delete(s.sessions, id)
			count++
		}
	}

	return count, nil
}

// Count returns the number of active sessions.
func (s *MemoryStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions), nil
}

// StartCleanup starts a goroutine that periodically cleans up expired sessions.
func (s *MemoryStore) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.Cleanup(ctx)
			}
		}
	}()
}
