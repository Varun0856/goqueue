package worker

import (
	"sync"
	"testing"

	"github.com/Varun0856/goqueue/internal/job"
	"github.com/Varun0856/goqueue/internal/store"
)

func TestPool_ConcurrentEnqueue(t *testing.T) {
	s := store.NewStore()
	pool := NewPool(8, 50, s)
	pool.Start()

	const numJobs = 50
	jobs := make([]*job.Job, numJobs)

	var wg sync.WaitGroup
	for i := 0; i < numJobs; i++ {
		j, err := job.New("uppercase", "payload")
		if err != nil {
			t.Fatalf("job.New: %v", err)
		}
		jobs[i] = j

		wg.Add(1)
		go func(j *job.Job) {
			defer wg.Done()
			if err := pool.Enqueue(j); err != nil {
				t.Errorf("failed to enqueue job %s: %v", j.ID, err)
			}
		}(j)
	}
	wg.Wait()
	pool.Stop()

	for _, j := range jobs {
		got, err := s.Get(j.ID)
		if err != nil {
			t.Fatalf("Get(%s): %v", j.ID, err)
		}
		if got.Status != job.Completed {
			t.Errorf("job %s: got status %q, want %q", j.ID, got.Status, job.Completed)
		}

		if got.Result == "" || got.Result != "PAYLOAD" {
			t.Errorf("job %s: unexpected result %v", j.ID, got.Result)
		}
	}
}

func TestPool_EnqueueAfterStop(t *testing.T) {
	s := store.NewStore()
	pool := NewPool(2, 10, s)
	pool.Start()
	pool.Stop()

	j, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.Enqueue(j)
	if err == nil {
		t.Fatalf("expected an error when enqueueing after Stop")
	}
}

func TestPool_StopDrainsQueue(t *testing.T) {
	s := store.NewStore()
	pool := NewPool(2, 10, s)
	pool.Start()

	const numJobs = 10

	jobs := make([]*job.Job, numJobs)

	for i := 0; i < numJobs; i++ {
		j, err := job.New("uppercase", "hello")
		if err != nil {
			t.Fatal(err)
		}
		jobs[i] = j

		if err := pool.Enqueue(j); err != nil {
			t.Fatalf("Enqueue: %v", err)
		}
	}
	pool.Stop()

	for _, j := range jobs {
		got, err := s.Get(j.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if got.Status != job.Completed {
			t.Errorf("job %s: expected completed, got %q", j.ID, got.Status)
		}
	}
}

func TestPool_StopMultipleTimes(t *testing.T) {
	pool := NewPool(2, 10, store.NewStore())
	pool.Start()

	pool.Stop()
	pool.Stop()
}

func TestPool_ConcurrentEnqueueAndStop(t *testing.T) {
	s := store.NewStore()
	pool := NewPool(2, 5, s)
	pool.Start()

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		j, err := job.New("uppercase", "hello")
		if err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
		go func(j *job.Job) {
			defer wg.Done()
			_ = pool.Enqueue(j)
		}(j)
	}
	pool.Stop()
	wg.Wait()
}

func TestPool_StartMultipleTimes(t *testing.T) {
	s := store.NewStore()
	pool := NewPool(3, 10, s)

	pool.Start()
	pool.Start()
	pool.Start()

	j, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatal(err)
	}

	if err := pool.Enqueue(j); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	pool.Stop()

	got, err := s.Get(j.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != job.Completed {
		t.Errorf("job %s: expected completed, got %q", j.ID, got.Status)
	}
}
