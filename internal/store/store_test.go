package store

import (
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
	s.Create(newJob)

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

	s.Create(newJob)

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

	s.Create(newJob)

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
