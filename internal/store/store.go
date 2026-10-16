package store

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/Varun0856/goqueue/internal/job"
)

type Store struct {
	mu   sync.RWMutex
	jobs map[string]job.Job
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]job.Job),
	}
}

func (s *Store) Create(j job.Job) {
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
	return existingJob, nil
}

func (s *Store) Update(j job.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.jobs[j.ID]
	if !ok {
		return errors.New("job not found")
	}

	s.jobs[j.ID] = j
	return nil
}

func (s *Store) List() []job.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]job.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	return jobs
}

func (s *Store) SaveToFile(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.jobs, "", " ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

func (s *Store) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var jobs map[string]job.Job

	if err := json.Unmarshal(data, &jobs); err != nil {
		return err
	}

	if jobs == nil {
		jobs = make(map[string]job.Job)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs = jobs
	return nil
}

func (s *Store) RecoverJob() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, j := range s.jobs {
		if j.Status == job.Running {
			j.Status = job.Queued
			s.jobs[id] = j
		}
	}
}
