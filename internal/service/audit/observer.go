package audit

import (
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

type Observer interface {
	Notify(e types.Event) error
	Close() error
}

type Subject struct {
	observers []Observer
	mu        sync.RWMutex
}

func (s *Subject) Register(o Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, o)
}

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

func (s *Subject) Notify(e types.Event) {
	s.mu.RLock()
	defer s.mu.Unlock()
	for _, observer := range s.observers {
		go observer.Notify(e)
	}
}
