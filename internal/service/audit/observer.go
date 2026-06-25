// Package audit provides observer pattern implementation for audit events
package audit

import (
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

// Observer interface defines the contract for audit event observers
// Implementing this interface allows objects to receive notifications about audit events
type Observer interface {
	// Notify sends an audit event to the observer
	// Returns error if notification fails
	Notify(e *types.Event) error
	// Close terminates the observer and releases resources
	// Returns error if closing fails
	Close() error
}

// Subject represents the audit event publisher
// It maintains a list of observers and notifies them of events
type Subject struct {
	observers []Observer
	mu        sync.RWMutex
}

// NewSubject creates and returns a new Subject instance
// Returns pointer to newly created Subject
func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

// Register adds an observer to the subject's list of observers
// Takes an Observer interface as parameter
func (s *Subject) Register(o Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, o)
}

// Remove removes an observer from the subject's list of observers
// Takes an Observer interface as parameter
func (s *Subject) Remove(o Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, observer := range s.observers {
		if observer == o {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			return
		}
	}
}

// Notify sends an audit event to all registered observers
// Takes an audit Event pointer as parameter
// Sends event to all observers concurrently using goroutines
func (s *Subject) Notify(e *types.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, observer := range s.observers {
		go observer.Notify(e)
	}
}
