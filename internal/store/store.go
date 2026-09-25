package store

import (
	"errors"
	"sync"

	"github.com/Varun0856/goqueue/internal/job"
)

type Store struct {
	mu   sync.RWMutex
	jobs map[string]*job.Job
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]*job.Job),
	}
}

func (s *Store) Create(j *job.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[j.ID] = j
}

func (s *Store) Get(id string) (job.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	existingJob, ok := s.jobs[id]
	if !ok {
		return job.Job{}, errors.New("job not found")
	}
	return *existingJob, nil
}

func (s *Store) Update(j job.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.jobs[j.ID]
	if !ok {
		return errors.New("job not found")
	}

	s.jobs[j.ID] = &j
	return nil
}
