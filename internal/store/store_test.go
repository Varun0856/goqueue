package store

import (
	"path/filepath"
	"testing"

	"github.com/Varun0856/goqueue/internal/job"
)

func TestCreateAndGet(t *testing.T) {
	// Create a Store
	s := NewStore()

	// Create a new job
	newJob, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("error while creating new job: %v", err)
	}

	// Store the job in Store
	s.Create(*newJob)

	// Retrieve the stored job
	storedJob, err := s.Get(newJob.ID)
	if err != nil {
		t.Fatalf("couldnt retrieve job: %v", err)
	}

	if newJob.ID != storedJob.ID {
		t.Errorf("expected %q, got %q", newJob.ID, storedJob.ID)
	}

	if newJob.Payload != storedJob.Payload {
		t.Errorf("expected %q, got %q", newJob.Payload, storedJob.Payload)
	}

	if newJob.Status != storedJob.Status {
		t.Errorf("expected %q, got %q", newJob.Status, storedJob.Status)
	}

	if newJob.Type != storedJob.Type {
		t.Errorf("expected %q, got %q", newJob.Type, storedJob.Type)
	}
}

func TestUpdate(t *testing.T) {
	s := NewStore()

	newJob, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("error while creating new job: %v", err)
	}

	s.Create(*newJob)

	storedJob, err := s.Get(newJob.ID)
	if err != nil {
		t.Fatalf("couldnt retrieve job: %v", err)
	}

	err = storedJob.Transition(job.Running)
	if err != nil {
		t.Fatalf("couldnt change the status of job: %v", err)
	}

	err = s.Update(storedJob)
	if err != nil {
		t.Fatalf("couldnt update the job in the store: %v", err)
	}

	updatedJob, err := s.Get(storedJob.ID)
	if err != nil {
		t.Fatalf("couldnt retrieve job: %v", err)
	}

	if updatedJob.Status != job.Running {
		t.Fatalf("expected status to be 'running' but got %q", updatedJob.Status)
	}
}

func TestGetNonExistentID(t *testing.T) {
	s := NewStore()

	_, err := s.Get("nonexistent-id")
	if err == nil {
		t.Fatalf("expected an error for a non-existent job id, but got nil")
	}
}

func TestUpdateNonExistentID(t *testing.T) {
	s := NewStore()

	newJob, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}

	err = s.Update(*newJob)
	if err == nil {
		t.Fatalf("expected an error for trying to update job which is not stored, but got nil")
	}
}

func TestGetCopy(t *testing.T) {
	s := NewStore()

	newJob, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}

	s.Create(*newJob)

	storedJob, err := s.Get(newJob.ID)
	if err != nil {
		t.Fatalf("failed to retrieve job: %v", err)
	}

	storedJob.Payload = "modified"

	updatedStoredJob, err := s.Get(storedJob.ID)
	if err != nil {
		t.Fatalf("failed to retrieve job: %v", err)
	}

	if updatedStoredJob.Payload != "hello" {
		t.Fatalf("expected stored payload to remain %q, got %q", "hello", updatedStoredJob.Payload)
	}

}

func TestStore_SaveAndLoad(t *testing.T) {
	s := NewStore()
	j, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}

	s.Create(*j)
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "store.json")

	if err := s.SaveToFile(filePath); err != nil {
		t.Fatalf("error while saving to store: %v", err)
	}

	newStore := NewStore()
	if err := newStore.LoadFromFile(filePath); err != nil {
		t.Fatalf("error while loading store: %v", err)
	}

	loadedJob, err := newStore.Get(j.ID)
	if err != nil {
		t.Fatalf("error while retrieving job: %v", err)
	}

	if loadedJob.ID != j.ID {
		t.Errorf("expected ID %s, got %s", j.ID, loadedJob.ID)
	}

	if loadedJob.Type != j.Type {
		t.Errorf("expected type %v, got %v", j.Type, loadedJob.Type)
	}

	if loadedJob.Payload != j.Payload {
		t.Errorf("expected payload %v, got %v", j.Payload, loadedJob.Payload)
	}

	if loadedJob.Status != j.Status {
		t.Errorf("expected status %v, got %v", j.Status, loadedJob.Status)
	}
}

func TestStore_SaveAndLoadRunningJobs(t *testing.T) {
	s := NewStore()
	j, err := job.New("uppercase", "hello world")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}
	s.Create(*j)

	storedJob, err := s.Get(j.ID)
	if err != nil {
		t.Fatalf("error while fetching the job: %v", err)
	}

	if err := storedJob.Transition(job.Running); err != nil {
		t.Fatalf("failed to transition job to Running: %v", err)
	}

	if err := s.Update(storedJob); err != nil {
		t.Fatalf("error while updating job: %v", err)
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "store.json")
	if err := s.SaveToFile(filePath); err != nil {
		t.Fatalf("error while saving to store: %v", err)
	}

	newStore := NewStore()
	if err := newStore.LoadFromFile(filePath); err != nil {
		t.Fatalf("error while loading store: %v", err)
	}
	newStoredJob, err := newStore.Get(j.ID)
	if err != nil {
		t.Fatalf("error while fetching job: %v", err)
	}
	if newStoredJob.Status != job.Running {
		t.Errorf("expected %v, got %v", job.Running, newStoredJob.Status)
	}

}

func TestStore_RecoverJobs(t *testing.T) {
	s := NewStore()

	j1, err := job.New("uppercase", "hello world")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}

	j2, err := job.New("uppercase", "hello")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}

	j3, err := job.New("uppercase", "world")
	if err != nil {
		t.Fatalf("error while creating a new job: %v", err)
	}

	s.Create(*j1)
	s.Create(*j2)
	s.Create(*j3)

	storedJob1, err := s.Get(j1.ID)
	if err != nil {
		t.Fatalf("error while retrieving job %s: %v", j1.ID, err)
	}

	storedJob2, err := s.Get(j2.ID)
	if err != nil {
		t.Fatalf("error while retrieving job %s: %v", j2.ID, err)
	}

	storedJob3, err := s.Get(j3.ID)
	if err != nil {
		t.Fatalf("error while retrieving job %s: %v", j3.ID, err)
	}

	// Job 1: Running
	if err := storedJob1.Transition(job.Running); err != nil {
		t.Fatalf("error transitioning job 1 to Running: %v", err)
	}

	// Job 2: Completed
	if err := storedJob2.Transition(job.Running); err != nil {
		t.Fatalf("error transitioning job 2 to Running: %v", err)
	}
	if err := storedJob2.Transition(job.Completed); err != nil {
		t.Fatalf("error transitioning job 2 to Completed: %v", err)
	}

	// Job 3: Failed
	if err := storedJob3.Transition(job.Running); err != nil {
		t.Fatalf("error transitioning job 3 to Running: %v", err)
	}
	if err := storedJob3.Transition(job.Failed); err != nil {
		t.Fatalf("error transitioning job 3 to Failed: %v", err)
	}

	// Update the store with the modified jobs.
	for _, j := range []job.Job{storedJob1, storedJob2, storedJob3} {
		if err := s.Update(j); err != nil {
			t.Fatalf("error updating job %s: %v", j.ID, err)
		}
	}

	// Recover jobs.
	s.RecoverJob()

	// Verify expected statuses.
	tests := []struct {
		name     string
		id       string
		expected job.JobStatus
	}{
		{"running job becomes queued", j1.ID, job.Queued},
		{"completed job remains completed", j2.ID, job.Completed},
		{"failed job remains failed", j3.ID, job.Failed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Get(tt.id)
			if err != nil {
				t.Fatalf("error retrieving job: %v", err)
			}

			if got.Status != tt.expected {
				t.Errorf("expected status %v, got %v", tt.expected, got.Status)
			}
		})
	}
}
