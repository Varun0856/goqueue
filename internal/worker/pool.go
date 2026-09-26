package worker

import (
	"errors"
	"log"
	"sync"

	"github.com/Varun0856/goqueue/internal/job"
	"github.com/Varun0856/goqueue/internal/store"
)

type Pool struct {
	jobs      chan string
	store     *store.Store
	workerNum int
	wg        sync.WaitGroup
	mu        sync.Mutex
	stopped   bool
	startOnce sync.Once
}

func NewPool(workerNum int, queueSize int, s *store.Store) *Pool {
	ch := make(chan string, queueSize)

	pool := Pool{
		jobs:      ch,
		store:     s,
		workerNum: workerNum,
	}
	return &pool
}

func (p *Pool) Enqueue(j *job.Job) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped {
		return errors.New("worker pool is stopped")
	}
	p.store.Create(*j)
	p.jobs <- j.ID
	return nil
}

func (p *Pool) Start() {
	p.startOnce.Do(func() {
		for i := 0; i < p.workerNum; i++ {
			p.wg.Add(1)
			go p.worker()
		}
	})
}

func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.stopped {
		p.stopped = true
		close(p.jobs)
	}
	p.mu.Unlock()
	p.wg.Wait()
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for id := range p.jobs {
		j, err := p.store.Get(id)
		if err != nil {
			log.Printf("worker: job %s not found in store: %v", id, err)
			continue
		}
		if err := j.Transition(job.Running); err != nil {
			log.Printf("failed to transition job %s to Running: %v", j.ID, err)
			continue
		}
		if err := p.store.Update(j); err != nil {
			log.Printf("failed to persist Running state for job %s: %v", j.ID, err)
			continue
		}

		result, err := Execute(j.Type, j.Payload)
		if err != nil {
			if transitionErr := j.Transition(job.Failed); transitionErr != nil {
				log.Printf("failed to transition job %s to Failed: %v", j.ID, transitionErr)
				continue
			}
			j.Error = err.Error()
			if updateErr := p.store.Update(j); updateErr != nil {
				log.Printf("failed to update failed job %s: %v", j.ID, updateErr)
			}
			continue
		}
		j.Result = result
		if err := j.Transition(job.Completed); err != nil {
			log.Printf("failed to transition job %s to Completed: %v", j.ID, err)
			continue
		}
		if err := p.store.Update(j); err != nil {
			log.Printf("failed to update completed job %s: %v", j.ID, err)
		}
	}
}
